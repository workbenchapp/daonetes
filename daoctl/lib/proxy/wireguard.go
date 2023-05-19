package proxy

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"

	"github.com/gagliardetto/solana-go"
	"github.com/go-logr/logr"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/memo"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const wireguardPublicPeerKeyName = "wgPeerKey"

// EnsureOnchainWireguardPeerKey checks if the on-chain wg-pubkey exists, or will put it on the chain.
func EnsureOnchainWireguardPeerKey(ctx context.Context, deviceAuthorityWallet *types.Account) string {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("Ensuring there is a wireguard peer key on-chain", "deviceAuthority", deviceAuthorityWallet.PublicKey)
	var memoInfo *memo.DaoletInfoMemo
	memoInfo, err := memo.GetFirstInfoMemo(ctx, deviceAuthorityWallet.PublicKey)

	if err == nil && memoInfo != nil {
		if _, ok := (*memoInfo)[wireguardPublicPeerKeyName]; !ok {
			// TODO: replace this with memo.GetFirstInfoMemoWithKey
			memoInfo = nil // not the right memo info
		}
	}

	dWgPrivateKey := getLocalDeviceWireguardPrivateKey(deviceAuthorityWallet)
	dWgPublicKey := dWgPrivateKey.PublicKey()

	if err != nil || memoInfo == nil { // TODO: or without peerKey?
		log.Info("No wireguard key found, making a new one")
		// OH wow - the config format needs the key in kex format, and that's not native to the wgtypes.Key
		publickeyInHex := hex.EncodeToString(dWgPublicKey[:])

		memoInfo = &memo.DaoletInfoMemo{
			wireguardPublicPeerKeyName: publickeyInHex,
		}
		memo.AddInfoMemo(ctx, memoInfo)
	}
	return dWgPublicKey.PublicKey().String()
}

// generateWireguardConfig generates both the cross-platform ipc config, and a wg-quick config
func generateWireguardConfig(ctx context.Context, dWgPrivateKey wgtypes.Key, deviceAuthorityWallet *types.Account, devices ProxyDeviceList) (string, string) {
	// etcWireguardConfig  is only for debugging - can be used with wg-quick to connect to the secret network
	log := logr.FromContextOrDiscard(ctx)
	// OH wow - the configuration-protocol format needs the key in kex format, and that's not native to the wgtypes.Key
	// TODO: need to add Address=localDevice.wireguardAddress
	etcWireguardConfig := fmt.Sprintf(`
[Interface]
PrivateKey=%s
ListenPort=12912`, dWgPrivateKey.String())
	// BUT the /etc/wireguard/wg0.conf format is different FFS (and for testing, i think i want to see both config formats)
	privatekeyInHex := hex.EncodeToString(dWgPrivateKey[:])

	config := fmt.Sprintf(`private_key=%s
listen_port=12912`, privatekeyInHex)
	// make https://www.wireguard.com/xplatform/#configuration-protocol
	// TODO: probably extract...
	var localDevice *ProxyDevice
	for deviceKey, device := range devices {
		if device.Info.DeviceAuthority.Equals(solana.PublicKey(deviceAuthorityWallet.PublicKey)) {
			// skip the local device
			localDevice = device
			// TODO: should update 				device.wireguardPeerKey = wgKey.String()
			continue
		}

		// TODO: use  127.1.0.x:12912 by default, and setup an ice.MakeNewICEConnectionRequest() for it (if there's no mDNS)
		//deviceAddr := net.IP(device.info.Ipv4[:]).String()
		deviceAddr := device.ProxyAddress + ":12913" // need to proxy on a different port

		// // Try mDns
		// if ip := QueryHostnameForIP(device.info.DeviceAuthority.String()); ip != nil {
		// 	deviceAddr = ip.String() + ":12912"
		// 	// TODO: check if there was a nat-traversal connection created, and drop it...
		// } else {
		// start the nat-traversal - internally, this will decide if it needs to create a new attempt, or just wait for the existing one
		//ice.MakeNewICEConnectionRequest(gOpts.Ctx, deviceKey+"Client", device.info.DeviceAuthority.String()+"Server", deviceAddr)
		// }
		// log.Info(
		// 	"Connecting to remote wireguard using:",
		// 	"deviceAddr", deviceAddr,
		// 	"deviceHostname", device.info.Hostname,
		// 	"deviceAuthority", device.info.DeviceAuthority.String(),
		// )

		// TODO: cache the peerKey
		devMemoInfo, err := memo.GetFirstInfoMemo(ctx, common.PublicKey(device.Info.DeviceAuthority))
		if err == nil && devMemoInfo != nil {
			if peerKey, ok := (*devMemoInfo)[wireguardPublicPeerKeyName]; ok {

				//AllowedIPs := "0.0.0.0/0"
				AllowedIPs := fmt.Sprintf("%s/32", device.WireguardAddress)

				keyInBytes, err := hex.DecodeString(peerKey)
				if err != nil {
					log.Error(err, "Failed to decode peerKey")
				}
				wgKey, err := wgtypes.NewKey(keyInBytes)
				if err != nil {
					log.Error(err, "Failed to get wireguard key")
				}
				device.WireguardPeerKey = wgKey.String()

				etcWireguardConfig = etcWireguardConfig + fmt.Sprintf(`
[Peer]
PublicKey=%s
Endpoint=%s
AllowedIPs=%s
PersistentKeepalive=25`, wgKey.String(), deviceAddr, AllowedIPs /*device.wireguardAddress*/)

				devConfig := fmt.Sprintf(`public_key=%s
endpoint=%s
allowed_ip=%s
persistent_keepalive_interval=25`, peerKey, deviceAddr, AllowedIPs /*device.wireguardAddress*/)
				config = config + "\n" + devConfig
			} else {
				setDeviceOff(deviceKey)
			}
		}
	}

	if localDevice == nil {
		log.Info("Device info not cached yet, skipping wg config")
		return "", ""
	}

	// TODO: Log this more elegantly
	log.V(1).Info("Wireguard config", "/etc/wireguard/wg0.conf", etcWireguardConfig)

	return config, localDevice.WireguardAddress
}

// TODO: by putting these into a struct, we can probably talk to more than one workgroup
var wireguardNet *netstack.Net
var wireguardDev *device.Device

// UpdateWireGuardNetwork should be triggered whenever a probable network topology change is detected
func UpdateWireGuardNetwork(ctx context.Context, deviceAuthorityWallet *types.Account, devices ProxyDeviceList) {
	log := logr.FromContextOrDiscard(ctx)
	if wireguardDev == nil {
		log.V(1).Info("Listen for ICEConnectionRequest")

		// TODO: really should make a wireguard service specific context, so we can cancel it, and start fresh.

		log.V(1).Info("Initializing wireguard network")
		initializeWireGuardNetwork(ctx, deviceAuthorityWallet, devices)
	} else {
		log.V(1).Info("Updating wireguard network")
		dWgPrivateKey := getLocalDeviceWireguardPrivateKey(deviceAuthorityWallet)
		// TODO: ARGH! devices is a global, stopit!
		wireguardConfig, wireguardAddress := generateWireguardConfig(ctx, dWgPrivateKey, deviceAuthorityWallet, devices)
		if wireguardConfig == "" && wireguardAddress == "" {
			return
		}

		log.V(1).Info("Wireguard config generated", "myIP", wireguardAddress, "wireguardConfig", wireguardConfig)

		// TODO: check if the config is differemt, if not, don't IpcSet..
		err := wireguardDev.IpcSet(wireguardConfig)
		if err != nil {
			panic(err)
		}
	}

}

func initializeWireGuardNetwork(ctx context.Context, deviceAuthorityWallet *types.Account, devices ProxyDeviceList) (*netstack.Net, error) {
	log := logr.FromContextOrDiscard(ctx)
	dWgPrivateKey := getLocalDeviceWireguardPrivateKey(deviceAuthorityWallet)
	wireguardConfig, wireguardAddress := generateWireguardConfig(ctx, dWgPrivateKey, deviceAuthorityWallet, devices)
	if wireguardConfig == "" && wireguardAddress == "" {
		return nil, fmt.Errorf("device info not cached yet, skipping wg config")
	}

	log.Info("Creating net TUN", "myIP", wireguardAddress, "wireguardConfig", wireguardConfig)

	tun, tnet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr(wireguardAddress)},
		[]netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("8.8.4.4")},
		1420,
	)
	if err != nil {
		panic(err)
	}
	logLevel := device.LogLevelError // TODO: set to verbose if -V
	//logLevel = device.LogLevelVerbose

	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(logLevel, "WG: "))
	// 	dev.IpcSet(`private_key=a8dac1d8a70a751f0f699fb14ba1cff7b79cf4fbd8f09f44c6e6a90d0369604f
	// listen_port=12912
	// public_key=25123c5dcd3328ff645e4f2a3fce0d754400d3887a0cb7c56f0267e20fbf3c5b
	// endpoint=163.172.161.0:12912
	// allowed_ip=0.0.0.0/0
	// persistent_keepalive_interval=25
	// `)
	err = dev.IpcSet(wireguardConfig)
	if err != nil {
		panic(err)
	}
	err = dev.Up()
	if err != nil {
		panic(err)
	}
	if err == nil {
		wireguardNet = tnet
		wireguardDev = dev
		go func() {
			<-ctx.Done()
			// TODO: OMG Don't ask (there's at least 40 goroutines that continue to exist if you don't close the wireguardDevice)
			// TODO: no utterly not goroutinesafe.
			wireguardDev.Close()
			wireguardDev = nil
			wireguardNet = nil
		}()
	}
	log.Info("Local device configured on wireguard", "wireguardAddress", wireguardAddress)

	return tnet, err
}

func getLocalDeviceWireguardPrivateKey(deviceAuthorityWallet *types.Account) wgtypes.Key {
	dWgPrivateKey, err := wgtypes.NewKey([]byte(deviceAuthorityWallet.PrivateKey)[:32])
	// TODO: log, and can we do something?
	if err != nil {
		panic(err)
	}
	return dWgPrivateKey
}

// This is the listener for the wireguard ports that should then request to the local deployment
func NOAddHttp(tnet *netstack.Net /*port, handler, idk*/) {
	listener, err := tnet.ListenTCP(&net.TCPAddr{Port: 9999})
	if err != nil {
		panic(err)
	}
	//mux := http.NewServeMux()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("> %s - %s - %s", request.RemoteAddr, request.URL.String(), request.UserAgent())
		io.WriteString(writer, "Hello from userspace TCP!")
	})
	//handler := cors.Default().Handler(mux)
	err = http.Serve(listener, nil)
	if err != nil {
		panic(err)
	}
}
