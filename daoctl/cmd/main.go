package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	gagliardetto "github.com/gagliardetto/solana-go"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/gops/agent"
	_ "github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/smartwallet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var cli struct {
	Verbose uint8 `help:"Log output Verbosity level." short:"v" default:"0"`
	Debug   bool  `help:"Extra debugging (enabled http://localhost:9495/debug/pprof)." default:"false"`

	Url               string `help:"Solana validator network name: [mainnet-beta, testnet, devnet, localhost]" short:"u" default:"devnet" yaml:"url"`
	GokiProgramPubkey string `help:"Public key for Goki smart wallet program" env:"GOKI_PROGRAM_PUBKEY" default:"GokivDYuQXPZCWRkwMhdH2h91KpDQXBEmpgBgs55bnpH"`

	// TODO: pretty sure ~ won't work on windows, need custom code
	KeyFile string `help:"File to load for payer key" yaml:"keypair" name:"keypair" default:"~/.config/solana/id.json" type:"path"`

	// TODO: It would be nice if this didn't have to be specified explicitly in daonetes.yaml. e.g., we could make it so
	// that when you add a user to a smart wallet, daonetes creates a PDA with a list of all the groups the user belongs
	// to or something like that.
	SmartWalletAddress string `help:"Address of the multisig smart wallet" yaml:"smart-wallet" name:"smart-wallet"`

	// TODO: Probably could just use index for this, then we'd be able to derive from smart wallet
	DerivedWalletAddress string `help:"Address of the smart wallet's SOL holding (i.e., transaction approving) wallet" yaml:"derived-wallet" name:"derived-wallet"`

	// Commandline options
	Agent DaoletCmd `cmd:"" help:"Run the Daolet agent (workload runner)"`
	//Deploy DeployCmd `cmd:"" help:"Manage deployments based on workspecs in the cluster"`
	//Device DeviceCmd `cmd:"" help:"Inspect and register devices on daonet"`
	Group GroupCmd `cmd:"" help:"Manage deployed networks"`
	//Spec SpecCmd `cmd:"" help:"Define workload specifications on daonet"`
	Expose ExposeCmd `cmd:"" help:"Expose a local port to the cluster"`
	Info   InfoCmd   `cmd:"" help:"Inspect daonet info"`

	// OS Service commands
	Status    StatusServiceCmd    `cmd:"" help:"Status of the Daolet agent OS Service"`
	Start     StartServiceCmd     `cmd:"" help:"Start the Daolet agent OS Service"`
	Stop      StopServiceCmd      `cmd:"" help:"Stop the Daolet agent OS Service"`
	Restart   RestartServiceCmd   `cmd:"" help:"Restart the Daolet agent OS Service"`
	Install   InstallServiceCmd   `cmd:"" help:"Install the Daolet agent OS Service"`
	Uninstall UninstallServiceCmd `cmd:"" help:"Uninstall the Daolet agent OS Service"`
	// Upgrade - should download, replace, restart - but this TODO: requires solig versioning

	Version VersionCmd `cmd:"" help:"Build version"`
}

func Main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	// TODO: I can't figure out a way to get this to work with the YAML keys having
	// snake_case names, without also forcing the flags to be --snake_case.
	// I don't like skewer-case keys in YAML, but it is what it is for now.
	cmdCtx := kong.Parse(&cli, kong.Configuration(kongyaml.Loader, "daonetes.yaml"))

	// TODO: need to allow url's, or other custom endpoints (but remember we do use the ws too)
	if c := program.GetClusterByName(cli.Url); c != nil {
		program.DefaultCluster = cli.Url
	}

	// setup Logging (chose this to setup for https://pkg.go.dev/go.opentelemetry.io/otel#SetLogger)
	zc := zap.NewDevelopmentConfig()
	zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zc.Level = zap.NewAtomicLevelAt(zapcore.Level(-cli.Verbose))
	z, err := zc.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialise logging: %v\n", err))
	}

	// setup our cancelable Context with ctrl-c
	osExitValue := -1
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM) // add SINGHUP to reload configs..

	go func() {
		<-ctx.Done()
		// do cleanup here
		os.Exit(osExitValue)
	}()
	go func() {
		// TODO: this needs to be a switch so we can do different things depending on context.
		// OR, use different channels...
		<-signalChan // first signal, cancel context
		signal.Stop(signalChan)
		z.Sugar().Info("Received interrupt, cleaning up...")
		// not sure what to set the code to when "legitimately" interrupted,
		// should send 130 (128+2) _if_ running interactively, but actually, we're mostly a service using Systemd etc
		// https://unix.stackexchange.com/questions/251996/why-does-bash-set-exit-status-to-non-zero-on-ctrl-c-or-ctrl-z
		osExitValue = 0
		cancel()
		<-signalChan // second ctrl-c for if the cleanup is hanging, hard exit
		os.Exit(-1)
	}()

	gokiPubkey := gagliardetto.MustPublicKeyFromBase58(cli.GokiProgramPubkey)
	smartwallet.SetProgramID(gokiPubkey)

	os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	os.Setenv("OTEL_EXPORTER_ENDPOINT", "https://api.honeycomb.io/")
	os.Setenv("OTEL_METRICS_ENABLED", "false")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://api.honeycomb.io/v1/metrics")
	os.Setenv("OTEL_SERVICE_NAME", "daoctl")
	// os.Setenv("HONEYCOMB_API_KEY", "yourkey")

	ctx = context.WithValue(ctx, options.URL, cli.Url)
	ctx = context.WithValue(ctx, options.GokiProgramPubkey, gokiPubkey)
	ctx = context.WithValue(ctx, options.KeyFile, cli.KeyFile)
	ctx = context.WithValue(ctx, options.SmartWalletAddress, cli.SmartWalletAddress)
	ctx = context.WithValue(ctx, options.DerivedWalletAddress, cli.DerivedWalletAddress)
	ctx = context.WithValue(ctx, options.Debug, cli.Debug)

	ourLog := zapr.NewLogger(z)
	ourCtx := logr.NewContext(ctx, ourLog)

	// Tracing setup for Honeycomb. Hack job to do this with env vars, but it's
	// quick.
	//
	// TODO: Port to calling the proper OpenTelemetry API and making all their
	// structs and stuff.
	//
	// os.Setenv("DEBUG", "true")

	gOpts := &options.GlobalOptions{
		Ctx: ourCtx,
		Log: ourLog,
	}

	go func() {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		markerReq, err := http.NewRequest(
			"POST",
			"https://api.honeycomb.io/1/markers/daoctl",
			bytes.NewBuffer([]byte(fmt.Sprintf(`{
				"message": "daoctl started on %s",
				"type": "process-start"
			}`, hostname))),
		)
		if err != nil {
			gOpts.Log.Error(err, "failed to create Honeycomb marker request")
			return
		}
		markerReq.Header.Set("X-Honeycomb-Team", "l9dHY17PtVTO7SJ9ejR6RB")
		if _, err := http.DefaultClient.Do(markerReq); err != nil {
			gOpts.Log.Error(err, "failed to Do Honeycomb marker request")
			/*
				// TODO: Erroneous
				if body, err := ioutil.ReadAll(resp.Body); err != nil {
					gOpts.Log.Error(err, "failed to read Honeycomb marker response body")
				} else {
					gOpts.Log.Info("Honeycomb marker response", "body", string(body))
				}
			*/
		}
	}()

	// Call the Run() method of the selected parsed command.
	err = cmdCtx.Run(gOpts)
	cmdCtx.FatalIfErrorf(err)
}
