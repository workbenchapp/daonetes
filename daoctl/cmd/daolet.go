package cmd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	bin "github.com/gagliardetto/binary"
	gagliardetto "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	gagliardettorpc "github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	gagliardettorws "github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/go-logr/logr"
	"github.com/honeycombio/opentelemetry-go-contrib/launcher"
	"github.com/portto/solana-go-sdk/types"
	serviceimpl "github.com/workbenchapp/worknet/daoctl/cmd/service"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/dns"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/ice"
	"github.com/workbenchapp/worknet/daoctl/lib/networking/pubip"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/proxy"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
	"github.com/workbenchapp/worknet/daoctl/lib/workgroup"
	"go.opentelemetry.io/otel"
)

const (
	MAX_REGISTRATION_RETRIES = 10
	MAX_FUNDING_RETRIES      = 30
	SLEEP_INTERVAL           = time.Second

	// from lib.rs
	DEVICE_ACCOUNT_SIZE = 128
)

type DaoletCmd struct {
	PollInterval  uint     `help:"Device deployment and Peer Poll interval in seconds" default:"120" yaml:"poll-interval"`
	ListenAddress string   `help:"Port to listen to for DAPP magic" default:"localhost:9495" yaml:"listenaddress"`
	FeatureFlags  []string `help:"Enable/Disable experimental features (disabledns|deployment)" default:"" yaml:"featureflags"`
	SignalServer  string   `help:"NAT busting connection negotiation service" default:"http://signal.daonetes.org:8080" yaml:"signalserver"`
}

type PrefixWriter struct {
	io.Writer
}

func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	go func() {
		for scanner.Scan() {
			if _, err := w.Write([]byte(prefix + scanner.Text() + "\n")); err != nil {
				panic(err)
			}
		}
	}()
	return &PrefixWriter{
		Writer: pipeWriter,
	}
}

func (pw *PrefixWriter) Write(p []byte) (n int, err error) {
	return pw.Writer.Write(p)
}

func downloadSpec(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func validateSpecChecksum(localSpecPath string, checksum string) error {
	specFile, err := os.Open(localSpecPath)
	if err != nil {
		return err
	}
	defer specFile.Close()

	specHash := sha256.New()
	if _, err := io.Copy(specHash, specFile); err != nil {
		return err
	}

	strHash := hex.EncodeToString(specHash.Sum(nil))
	if strHash != checksum {
		return fmt.Errorf(
			"checksum mismatch with spec: expected=%s actual=%s",
			strHash,
			checksum,
		)
	}

	return nil
}

func (r *DaoletCmd) featureFlagEnabled(flag string) bool {
	for _, f := range r.FeatureFlags {
		if f == flag {
			return true
		}
	}
	return false
}

func (r *DaoletCmd) Run(gOpts *options.GlobalOptions) error {
	parentCtx := gOpts.Ctx
	ctx, cancel := context.WithCancel(parentCtx)
	gOpts.Ctx = ctx
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP) // add SINGHUP to reload configs..
	go func() {
		for {

			// TODO: this needs to be a switch so we can do different things depending on context.
			// OR, use different channels...
			s := <-signalChan // first signal, cancel context
			//signal.Stop(signalChan) // Don't stop the signals
			gOpts.Log.Info("Received Signal restarting", "signal", s.String())
			// not sure what to set the code to when "legitimately" interrupted,
			// should send 130 (128+2) _if_ running interactively, but actually, we're mostly a service using Systemd etc
			// https://unix.stackexchange.com/questions/251996/why-does-bash-set-exit-status-to-non-zero-on-ctrl-c-or-ctrl-z
			cancel()
		}
	}()

	var err error
	for {
		gOpts.Log.Info("RestartableRun loop")

		workgroup.InitDeviceCache() // start with a fresh info cache

		err = r.RestartableRun(gOpts)
		if err != nil {
			gOpts.Log.Info("RestartableRun break", "err", err)
			cancel()
			//break

			gOpts.Log.Info("WAIT 10s to restart loop") // TODO: this is to allow things to stop and for debugging (tune downwards)
			<-time.After(time.Duration(10) * time.Second)
			gOpts.Log.Info("timed out, restarting")

		}

		// reset the context to go again
		ctx, cancel = context.WithCancel(parentCtx)
		gOpts.Ctx = ctx
	}
	return err
}

func (r *DaoletCmd) RestartableRun(gOpts *options.GlobalOptions) error {
	ctx := gOpts.Ctx

	otelShutdown, err := launcher.ConfigureOpenTelemetry()
	if err != nil {
		gOpts.Log.Error(err, "error setting up OTel SDK")
	}
	defer otelShutdown()

	ctx = context.WithValue(ctx, options.TracerKey, otel.Tracer("daoctl"))

	gOpts.Log.Info(
		"Agent starting",
		"poll_interval", r.PollInterval,
		"listen_address", r.ListenAddress,
		"signalserver", r.SignalServer,
	)

	// TODO: test that ListenAddress is valid...
	// TODO: add some way to disable the web service - we'll still proxy things, but we're more secretive about it.
	// TODO: Sven claims this is essentially safe, as its the same as the data on the chain
	// BUT - its a lie, this endpoint confirms that this is a specific account on chain
	proxy.ListenAndServeLocalhost(ctx, r.ListenAddress)
	if !serviceimpl.Admin() {
		gOpts.Log.Info(".dmesh DNS disabled, not running as root/Admin")
	} else {
		if !r.featureFlagEnabled("disabledns") {
			// TODO: test if we have permission to listen on port 53
			err := dns.EnsureDNSConfigured()
			if err != nil {
				panic(err)
			}
			go dns.RunDnsService(ctx)
		}
	}

	agentConfig, err := options.Config()
	if err != nil {
		return fmt.Errorf("error getting or creating agent config: %s", err)
	}

	gOpts.Log.Info("Starting new mesh", "mesh name", agentConfig.ActiveNet)

	activeNet, err := agentConfig.Active()
	if err != nil {
		return err
	}

	ourWallet, err := solana.MustGetAccount(ctx, "WorkNet", activeNet.KeyFile)
	if err != nil {
		return err
	}

	seeds := [][]byte{
		ourWallet.PublicKey.Bytes(),
	}

	// Make it easier to find any devices that are on our local network
	go proxy.ResolveMDNS(ctx)
	go proxy.ServeMDNS(ctx, ourWallet.PublicKey.String())

	deviceInfoKey, deviceBump, err := gagliardetto.FindProgramAddress(seeds, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return err
	}

	gOpts.Log.Info("Looking for PDA", "devicePDA", deviceInfoKey)

	client := gagliardettorpc.New(options.SolanaCluster(ctx).RPC)

	var deviceAccountResp *gagliardettorpc.GetAccountInfoResult
	for {
		deviceAccountResp, err = client.GetAccountInfo(ctx, deviceInfoKey)
		if err == nil {
			break
		}
		if err == gagliardettorpc.ErrNotFound {
			err = errors.New("no PDA found. Must register device:\ndaoctl device register " + ourWallet.PublicKey.String())
		}
		gOpts.Log.Info("Getting local device account info", "state", err.Error())

		select {
		case <-ctx.Done():
			fmt.Println("context canceled")
			return fmt.Errorf("Main agent loop context canceled")
		case <-time.After(time.Duration(r.PollInterval) * time.Second):
			fmt.Println("timed out")
		}
	}

	gOpts.Log.Info("Device account found on chain", "devicePDA", deviceInfoKey.String())

	device := &worknet.Device{}
	deviceAccount := deviceAccountResp.Value
	decoder := bin.NewDecoderWithEncoding(deviceAccount.Data.GetBinary(), bin.EncodingBorsh)
	if err := device.UnmarshalWithDecoder(decoder); err != nil {
		return err
	}

	if device.Status == worknet.DeviceStatusRegistrationRequested {
		if err := RegisterDevice(gOpts, client, device, ourWallet, deviceInfoKey, deviceBump); err != nil {
			return err
		}
	}

	// Prime the cache so that the ProxyToDevices has something to look at
	// TODO: this is dumb :) - need to work out how we refresh this...
	workgroup.GetDeviceInfo(ctx)
	// TODO: this should be integrated into the device chain metadata
	/*myWireguardPublicKey :=*/
	proxy.EnsureOnchainWireguardPeerKey(ctx, ourWallet)
	gOpts.Ctx = context.WithValue(ctx, ice.GetSignalServerContextKey, r.SignalServer)
	go ice.ListenForICEConnectionRequest(ctx, ourWallet.PublicKey.String()+"Server", "127.0.0.1:12912")
	// Cool, we're ready to accept work, LFG
	for {
		// TODO: want to make one polling system that only requests data from the chain or its peers
		// TODO: and everything else listens to see if the cached info means it needs to act.

		proxy.ProxyToDevices(ctx, ourWallet, r.ListenAddress) // TODO: so this should probably move to its own event system

		if err := r.UpdateDeployments(ctx, client, device, ourWallet, activeNet); err != nil {
			gOpts.Log.Error(err, "Update loop", "client", client, "device", device, "wallet", ourWallet)
		}

		// TODO: we should have WS subscription(s) instead of polling
		//time.Sleep(time.Duration(r.PollInterval) * time.Second)
		fmt.Println("main agent loop")
		select {
		case <-ctx.Done():
			fmt.Println("context canceled")
			return fmt.Errorf("Main agent loop context canceled")
		case <-time.After(time.Duration(r.PollInterval) * time.Second):
			fmt.Println("timed out")
		}
	}
}

func (r *DaoletCmd) UpdateDeployments(
	ctx context.Context,
	client *gagliardettorpc.Client,
	device *worknet.Device,
	ourWallet *types.Account,
	worknetCfg *options.WorknetConfig,
	//	deviceInfoKey gagliardetto.PublicKey,
	//	deviceBump uint8,
) error {
	devicePubKey := gagliardetto.PublicKeyFromBytes(ourWallet.PublicKey.Bytes())
	log := logr.FromContextOrDiscard(ctx)

	deviceDeployTokenAccounts, err := client.GetTokenAccountsByOwner(
		ctx,
		devicePubKey,
		&gagliardettorpc.GetTokenAccountsConfig{
			ProgramId: &gagliardetto.TokenProgramID,
		},
		&gagliardettorpc.GetTokenAccountsOpts{},
	)
	if err != nil {
		return err
	}

	for _, tokenAccount := range deviceDeployTokenAccounts.Value {
		// TODO: these "return err" should also be "continue" - maybe extract to function?
		// TODO: or better yet, move to "device" module, and then have a "deployment" module?
		tokenWallet := &token.Account{}
		decoder := bin.NewDecoderWithEncoding(tokenAccount.Account.Data.GetBinary(), bin.EncodingBorsh)
		if err := tokenWallet.UnmarshalWithDecoder(decoder); err != nil {
			return fmt.Errorf("couldn't decode token account: %s", err)
		}

		deploymentMintAccountInfoResp, err := client.GetAccountInfo(ctx, tokenWallet.Mint)
		if err != nil {
			return fmt.Errorf("couldn't get deployment mint account: %s", err)
		}

		mint := &token.Mint{}
		decoder = bin.NewDecoderWithEncoding(
			deploymentMintAccountInfoResp.Value.Data.GetBinary(),
			bin.EncodingBorsh,
		)
		if err := mint.UnmarshalWithDecoder(decoder); err != nil {
			return fmt.Errorf("couldn't decode deployment mint account: %s", err)
		}

		deploymentAccountInfoResp, err := client.GetAccountInfo(ctx, *mint.MintAuthority)
		if err != nil {
			return fmt.Errorf("couldn't get deployment mint account: %s", err)
		}

		deployment := &worknet.Deployment{}
		decoder = bin.NewDecoderWithEncoding(
			deploymentAccountInfoResp.Value.Data.GetBinary(),
			bin.EncodingBorsh,
		)
		if err := deployment.UnmarshalWithDecoder(decoder); err != nil {
			return fmt.Errorf("couldn't decode deployment account: %s", err)
		}

		// device is getting group authority
		// not work_group account
		//
		// off because it's the identifier and not the full
		// group authority any more
		deploymentPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
			device.WorkGroup.Bytes(),
			[]byte(deployment.Name),
			[]byte("deployment"),
		}, program.WORKNET_V1_PROGRAM_PUBKEY)
		if err != nil {
			return fmt.Errorf("couldn't get deployment PDA: %s", err)
		}

		// make sure deployment and deployment_mint are issued from
		// the group authority in question, i.e., that we don't
		// run workloads from a token account passed to us arbitrarily
		//
		// TODO: need to verify this cryptographically
		if deploymentPDA.String() != mint.MintAuthority.String() {
			log.Error(err, "Invalid deployment token account, skipping",
				"deployment.Name", deployment.Name,
				"deploymentPDA", deploymentPDA,
				"tokenAccount", tokenAccount.Pubkey,
				"tokenWallet.Mint", tokenWallet.Mint,
				"mint.MintAuthority", mint.MintAuthority,
			)
			continue
		}

		specAccountInfoResp, err := client.GetAccountInfo(ctx, deployment.Spec)
		if err != nil {
			return fmt.Errorf("couldn't get spec account: %s", err)
		}

		spec := &worknet.WorkSpec{}
		decoder = bin.NewDecoderWithEncoding(
			specAccountInfoResp.Value.Data.GetBinary(),
			bin.EncodingBorsh,
		)
		if err := spec.UnmarshalWithDecoder(decoder); err != nil {
			return fmt.Errorf("couldn't decode spec account: %s", err)
		}
		err = r.updateDeployment(ctx, spec, deployment, deploymentPDA)
		if err != nil {
			return fmt.Errorf("error updating deployment: %s", err)
		}
	}

	workgroup.UpdateDeployState(ctx, "", "local", workgroup.DeploymentInfo{
		Deployment: worknet.Deployment{},
		Spec:       worknet.WorkSpec{},
		States: []workgroup.DeployState{
			{Publishers: worknetCfg.Ports},
		},
	})

	return nil
}

func saveState(ctx context.Context, deploymentPDA gagliardetto.PublicKey, spec *worknet.WorkSpec, deployment *worknet.Deployment, scheduleWorkDirPath, localSpecPath, projectName string) {
	log := logr.FromContextOrDiscard(ctx)
	specPath := filepath.Join(scheduleWorkDirPath, "spec.json")
	// TODO: skip if already written
	// TODO: log if the onchain spec has changed...
	specJSON, _ := json.MarshalIndent(*spec, "", " ") // TODO: json err...
	if err := ioutil.WriteFile(specPath, specJSON, 0644); err != nil {
		log.Error(err,
			"Could not write spec JSON",
			"specPath", specPath,
		)
	}
	deploymentPath := filepath.Join(scheduleWorkDirPath, "deployment.json")
	// TODO: skip if already written
	// TODO: log if the onchain spec has changed...
	deploymentJSON, _ := json.MarshalIndent(*deployment, "", " ") // TODO: json err...
	if err := ioutil.WriteFile(deploymentPath, deploymentJSON, 0644); err != nil {
		log.Error(err,
			"Could not write deployment JSON",
			"deploymentPath", deploymentPath,
		)
	}
	// And now get the compose state
	statePath := filepath.Join(scheduleWorkDirPath, "state.json")
	stateCmd := exec.Command(
		"docker-compose", "--project-name", projectName,
		"--file", localSpecPath, "ps", "--all", "--format", "json",
	)

	stateBytes, err := stateCmd.Output()
	if err != nil {
		log.Error(err,
			"Getting compose ps failed",
			"projectName",
			projectName,
		)
	}

	states := []workgroup.DeployState{}
	json.Unmarshal(stateBytes, &states)                 // TODO: json err...
	stateJSON, _ := json.MarshalIndent(states, "", " ") // TODO: json err...
	if err = ioutil.WriteFile(statePath, stateJSON, 0644); err != nil {
		log.Error(err,
			"Could not write state JSON",
			"statePath", statePath,
		)
	}
	// TODO: how do i get the error...

	workgroup.UpdateDeployState(ctx, "", deploymentPDA.String(), workgroup.DeploymentInfo{
		Deployment: *deployment,
		Spec:       *spec,
		States:     states,
	})
}

func (r *DaoletCmd) updateDeployment(ctx context.Context, spec *worknet.WorkSpec, deployment *worknet.Deployment, deploymentPDA gagliardetto.PublicKey) error {
	log := logr.FromContextOrDiscard(ctx)
	// compose/swarm doesn't like long names
	projectName := "daonetes" + strings.ToLower(deploymentPDA.String()[:16])

	specWorkDirsPath := "specworkdirs"
	if _, err := os.Stat(specWorkDirsPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(specWorkDirsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// TODO: this means that an agent can only have one deployment of a specific spec
	//       what happens if there is more than on on a device? or even 2 replicas?
	//       Sven suspects should at least be based on the deployment token ata (so that _if_ someone starts 2 different agents from the same dir, we don't clash)
	scheduleWorkDirPath := filepath.Join(
		specWorkDirsPath,
		deployment.Spec.String(),
	)

	if _, err := os.Stat(scheduleWorkDirPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(scheduleWorkDirPath, os.ModePerm)
		if err != nil {
			log.Error(err, "Couldn't make directory", "dirpath", scheduleWorkDirPath)
		}
	}

	localSpecPath := filepath.Join(scheduleWorkDirPath, "spec")
	if strings.HasPrefix(spec.UrlOrContents, "https://") {
		splitURL := strings.Split(spec.UrlOrContents, "/")

		// tack on "-docker-compose.yaml", etc
		localSpecPath += "-" + splitURL[len(splitURL)-1]
	} else {
		// contents...
		// TODO: rename this to the right type, based on spec.workType
	}

	// TODO: save where updateDeployment got up to too
	defer saveState(ctx, deploymentPDA, spec, deployment, scheduleWorkDirPath, localSpecPath, projectName)

	if stat, err := os.Stat(localSpecPath); err == nil {
		// TODO: at this point, this doesn't point at the downloaded filename ...
		specLastModified := time.Unix(int64(spec.ModifiedAt), 0)

		// TODO We might want to periodically re-apply the spec,
		// or force an update with some type of agent --force-update
		// flag. Will have to revisit + check libcompose code to see
		// how elegantly it diffs.
		// For now, you can force it by deleteing the spec file
		if stat.ModTime().After(specLastModified) {
			return err
		}
	}
	if r.featureFlagEnabled("deployment") {
		// Don't start / stop things that are deployed.
		// TODO: can we check if it's deployed / running / dead? (and is knowing that useful?)

		// TODO: download, then check checksum, if its incorrect, don't overwrite the old spec, (cos we do want to know what _was_ deployed, not what wasn't)
		if strings.HasPrefix(spec.UrlOrContents, "https://") {
			if err := downloadSpec(localSpecPath, spec.UrlOrContents); err != nil {
				log.Error(err,
					"Could not download spec",
					"specURL", spec.UrlOrContents,
					"specPDA", deployment.Spec.String(),
					err,
				)
				return err
			}
		} else {
			if err := ioutil.WriteFile(localSpecPath, []byte(spec.UrlOrContents), 0644); err != nil {
				log.Error(err, "Could not write spec contents", "specPDA", deployment.Spec.String())
				return err
			}
		}

		if err := validateSpecChecksum(localSpecPath, spec.ContentsSha256); err != nil {
			log.Error(err,
				"Could not validate spec checksum",
				"specPDA",
				deployment.Spec.String(),
			)
			return err
		}

		if spec.WorkType == worknet.WorkTypeDockerCompose {
			log.Info("Docker Compose worknet spec", "specName", spec.Name)

			// Groan, stack deploy is a magic abstraction that
			// doesn't exist first class in the Docker API. See
			// https://stackoverflow.com/questions/42155978/docker-stack-deploy-using-the-client-api
			//
			// So we just send it and exec for now.
			// TODO: Be more elegant.
			deployCmd := exec.Command(
				"docker-compose", "--project-name", projectName,
				"--file", localSpecPath, "up", "-d",
			)
			deployCmd.Stdout = NewPrefixWriter(os.Stdout, "DOCKEROUT => ")
			deployCmd.Stderr = NewPrefixWriter(os.Stderr, "DOCKERERR => ")
			log.Info("Deploying Compose spec", "specName", spec.Name, "specPath", localSpecPath)
			if err := deployCmd.Run(); err != nil {
				log.Error(err, "Deploying spec failed", "specName", spec.Name, "specPDA", deployment.Spec.String())
				return err
			}
		} else {
			log.Info("Unknown worknet spec", "spec", spec)
		}
	}
	return nil
}

func RegisterDevice(
	gOpts *options.GlobalOptions,
	client *gagliardettorpc.Client,
	device *worknet.Device,
	ourWallet *types.Account,
	deviceInfoKey gagliardetto.PublicKey,
	deviceBump uint8,
) error {
	gOpts.Log.Info("Polling for device announcement funding")

	wsClient, err := gagliardettorws.Connect(gOpts.Ctx, options.SolanaCluster(gOpts.Ctx).WS)
	if err != nil {
		return err
	}
	devicePubKey := gagliardetto.PublicKeyFromBytes(ourWallet.PublicKey.Bytes())

	// var deviceAuthorityAccountInfo gagliardettorpc.Account

	// Poll to see if we got our "seed" funding to announce
	// our properties (e.g., IPv4) to the network.
	//
	// TODO: This could probably just be a websocket programChanges
	// subscription.
	for i := 0; i < MAX_FUNDING_RETRIES; i++ {
		if i == MAX_FUNDING_RETRIES-1 {
			return errors.New("timed out waiting for funding")
		}
		if _, err := client.GetAccountInfo(gOpts.Ctx, device.DeviceAuthority); err != nil {
			if err == gagliardettorpc.ErrNotFound {
				time.Sleep(SLEEP_INTERVAL)
				continue
			}
			return fmt.Errorf("error while polling for announcement funding: %s", err)
		} else {
			// deviceAuthorityAccountInfo = *resp.Value
			break
		}
	}

	// TODO: check deposited lamports enough to pay for txn
	gOpts.Log.Info("Funding obtained, announcing device to network")

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	ip, err := pubip.IP()
	if err != nil {
		return err
	}
	gOpts.Log.Info(
		"Announcing device to worknet",
		"device_pubkey", devicePubKey,
		"device_hostname", hostname,
		"device_ipv4", ip,
	)
	ipv4Bytes := []byte(ip.To4())
	updateDeviceInst, err := worknet.NewUpdateDeviceInstruction(
		[4]byte{ipv4Bytes[0], ipv4Bytes[1], ipv4Bytes[2], ipv4Bytes[3]},
		strings.ToLower(hostname),
		deviceBump,
		worknet.DeviceStatusRegistered,
		devicePubKey,
		deviceInfoKey,
	).ValidateAndBuild()
	if err != nil {
		return err
	}

	/*
		// leave out for now -- sol wants rent exempt minimum even on empty accounts >:-|
		refundTxn, err := system.NewTransferInstruction(
			deviceAuthorityAccountInfo.Lamports-5000,
			devicePubKey,
			device.GroupAuthority,
		).ValidateAndBuild()
		if err != nil {
			return err
		}
	*/
	blockHash, err := client.GetRecentBlockhash(gOpts.Ctx, gagliardettorpc.CommitmentFinalized)
	if err != nil {
		return err
	}

	updateDeviceTxn, err := gagliardetto.NewTransaction(
		[]gagliardetto.Instruction{
			updateDeviceInst,
			// refundTxn,
		},
		blockHash.Value.Blockhash,
	)
	if err != nil {
		return err
	}

	_, err = updateDeviceTxn.Sign(func(key gagliardetto.PublicKey) *gagliardetto.PrivateKey {
		return (*gagliardetto.PrivateKey)(&ourWallet.PrivateKey)
	})
	if err != nil {
		return err
	}

	_, err = sendandconfirmtransaction.SendAndConfirmTransaction(
		gOpts.Ctx,
		client,
		wsClient,
		updateDeviceTxn,
	)
	if err != nil {
		return err
	}

	return nil
}
