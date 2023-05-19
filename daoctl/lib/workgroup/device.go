package workgroup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	bin "github.com/gagliardetto/binary"
	gagliardetto "github.com/gagliardetto/solana-go"
	gagliardettorpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/go-logr/logr"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
	"github.com/workbenchapp/worknet/daoctl/lib/version"
)

type DeployState struct {
	Publishers []options.Publisher
}

type DeploymentInfo struct {
	Deployment worknet.Deployment `json:"deployment"`
	Spec       worknet.WorkSpec   `json:"spec"`
	States     []DeployState      `json:"state"`
}

type DeviceStatusInfo struct {
	DeployState      map[string]DeploymentInfo `json:"deploy_state"`
	DeviceInfo       worknet.Device            `json:"deviceInfo"`
	GroupInfo        worknet.WorkGroup         `json:"groupInfo"`
	DeviceInfoKey    string                    `json:"deviceInfoKey"`
	DeviceInfoKeBump string                    `json:"deviceInfoKeyBump"`
	DeviceWallet     string                    `json:"deviceWallet"`
	Validator        string                    `json:"validator"`

	// Agent code version
	Version         string
	VersionRevision string
	VersionDate     string
}

// TODO: all this hacked up BS cache grew from a POC POS, and needs to be replaced by a mesh gossip cache (ask sven, there was a plan)

// the remote status cache (keyd by deviceATA)
var remoteDeviceCache sync.Map // map[string]*DeviceStatusInfo
func InitDeviceCache() {
	remoteDeviceCache = sync.Map{} // map[string]*DeviceStatusInfo
}

func UpdateDeployState(ctx context.Context, deviceATA, deployKey string, data DeploymentInfo) {
	log := logr.FromContextOrDiscard(ctx)
	// TODO: so not threadsafe
	if deviceATA == "" {
		deviceATA = "local"
	}
	log.V(1).Info("Caching deploy state", "deviceTokenAccount", deviceATA)

	currentInfo := &DeviceStatusInfo{}
	loadCurrentInfo, ok := remoteDeviceCache.Load(deviceATA)
	if ok {
		currentInfo = loadCurrentInfo.(*DeviceStatusInfo)
	}
	if currentInfo.DeployState == nil {
		currentInfo.DeployState = make(map[string]DeploymentInfo)
	}
	currentInfo.DeployState[deployKey] = data
	remoteDeviceCache.Store(deviceATA, currentInfo)
}

// from the proxy requests...
// this is a horrifying result of trying to avoid making too many requests to the chain
func UpdateDeviceStatusInfo(ctx context.Context, data []byte) {
	log := logr.FromContextOrDiscard(ctx)
	// TODO: so not threadsafe
	var currentInfo DeviceStatusInfo
	err := json.Unmarshal(data, &currentInfo)
	if err != nil {
		log.Error(err, "Couldn't unmarshal device status info")
		//spew.Dump(data)
		return
	}

	log.V(1).Info("Caching UpdateDeviceStatue for current device", "deviceAuthority", currentInfo.DeviceInfo.DeviceAuthority.String())

	remoteDeviceCache.Store(currentInfo.DeviceInfo.DeviceAuthority.String(), &currentInfo)
}

// TODO: yeah.
func GetCachedDeviceStatusInfo(deviceATA string) *DeviceStatusInfo {
	if deviceATA == "" {
		deviceATA = "local"
	}
	status, ok := remoteDeviceCache.Load(deviceATA)
	if !ok {
		return nil
	}

	return status.(*DeviceStatusInfo)
}

func GetCachedWorkGroupInfo() *worknet.WorkGroup {
	deviceATA := "local"

	status, ok := remoteDeviceCache.Load(deviceATA)
	if !ok {
		return nil
	}
	return &status.(*DeviceStatusInfo).GroupInfo
}

func GetDeviceInfo(ctx context.Context) (*DeviceStatusInfo, error) {
	deviceATA := "local"
	currentInfo := &DeviceStatusInfo{}
	loadCurrentInfo, ok := remoteDeviceCache.Load(deviceATA)
	if ok {
		currentInfo = loadCurrentInfo.(*DeviceStatusInfo)
	}
	defer func() {
		// cache whatever info we got...
		remoteDeviceCache.Store(deviceATA, currentInfo)
	}()

	ourWallet, err := solana.MustGetAgentWallet(ctx)
	if err != nil {
		return nil, err
	}

	currentInfo.DeviceWallet = ourWallet.PublicKey.String()
	currentInfo.Version = version.GetVersionString()
	currentInfo.VersionRevision = version.GetBuildRevision()
	currentInfo.VersionDate = version.GetBuildDate()

	seeds := [][]byte{
		ourWallet.PublicKey.Bytes(),
	}

	// convert from solana PublicKey to gagliardetto PublicKey :/
	// devicePubKey := gagliardetto.PublicKeyFromBytes(ourWallet.PublicKey.Bytes())

	// device info
	key, deviceBump, err := gagliardetto.FindProgramAddress(seeds, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return currentInfo, err
	}

	currentInfo.DeviceInfoKey = key.String()
	currentInfo.DeviceInfoKeBump = fmt.Sprintf("%v", deviceBump)

	currentInfo.Validator = options.SolanaCluster(ctx).RPC
	client := gagliardettorpc.New(options.SolanaCluster(ctx).RPC)
	// wsClient, err := gagliardettorws.Connect(gOpts.Ctx, options.SolanaCluster(gOpts.Ctx).WS)
	// if err != nil {
	// 	return result, err
	// }

	var deviceAccountResp *gagliardettorpc.GetAccountInfoResult

	if deviceAccountResp, err = client.GetAccountInfo(ctx, key); err != nil {
		if err == gagliardettorpc.ErrNotFound {
			return currentInfo, errors.New("no PDA found. Must register device:\ndaoctl device register " + ourWallet.PublicKey.String())
		} else {
			return currentInfo, err
		}
	}

	device := &worknet.Device{}
	deviceAccount := deviceAccountResp.Value
	decoder := bin.NewDecoderWithEncoding(deviceAccount.Data.GetBinary(), bin.EncodingBorsh)
	if err := device.UnmarshalWithDecoder(decoder); err != nil {
		return currentInfo, err
	}
	currentInfo.DeviceInfo = *device
	// CAN test for pre-14August2022 device by seeing what program type the "WorkGroup" entry in deviceInfo is - if its Solana SystemProgram, then its legacy
	var groupAccountResp *gagliardettorpc.GetAccountInfoResult

	if groupAccountResp, err = client.GetAccountInfo(ctx, device.WorkGroup); err != nil {
		if err == gagliardettorpc.ErrNotFound {
			return currentInfo, errors.New("no PDA found. Must register device:\ndaoctl device register " + ourWallet.PublicKey.String())
		} else {
			return currentInfo, err
		}
	}

	group := &worknet.WorkGroup{}
	groupAccount := groupAccountResp.Value
	groupDecoder := bin.NewDecoderWithEncoding(groupAccount.Data.GetBinary(), bin.EncodingBorsh)
	if err := group.UnmarshalWithDecoder(groupDecoder); err != nil {
		return currentInfo, errors.New("group Key (" + device.WorkGroup.String() + ") doesn't point to a current workgroup account: " + err.Error())
	}
	currentInfo.GroupInfo = *group

	return currentInfo, nil
}

func GetDeviceInfoByKey(ctx context.Context, deviceKey gagliardetto.PublicKey) (info *worknet.Device, err error) {
	client := gagliardettorpc.New(options.SolanaCluster(ctx).RPC)
	// wsClient, err := gagliardettorws.Connect(gOpts.Ctx, options.SolanaCluster(gOpts.Ctx).WS)
	// if err != nil {
	// 	return result, err
	// }

	var deviceAccountResp *gagliardettorpc.GetAccountInfoResult

	if deviceAccountResp, err = client.GetAccountInfo(ctx, deviceKey); err != nil {
		if err == gagliardettorpc.ErrNotFound {
			return nil, errors.New("device " + deviceKey.String() + " not found on chain")
		} else {
			return nil, err
		}
	}

	device := &worknet.Device{}
	deviceAccount := deviceAccountResp.Value
	decoder := bin.NewDecoderWithEncoding(deviceAccount.Data.GetBinary(), bin.EncodingBorsh)
	if err := device.UnmarshalWithDecoder(decoder); err != nil {
		return nil, err
	}
	return device, nil
}
