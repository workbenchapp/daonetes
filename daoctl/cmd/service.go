package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-logr/zapr"
	"github.com/kardianos/service"
	serviceimpl "github.com/workbenchapp/worknet/daoctl/cmd/service"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StatusServiceCmd struct{}

func (r *StatusServiceCmd) Run(gOpts *options.GlobalOptions) error {
	s, err := common()
	if err != nil {
		return fmt.Errorf("common: %s", err)
	}

	status, err := s.Status()
	if err != nil {
		return fmt.Errorf("get status error: %s", err)
	}
	stateString := "possibly not installed"
	switch status {
	case service.StatusUnknown:
	case service.StatusRunning:
		stateString = "running"
	case service.StatusStopped:
		stateString = "stopped"

	}
	fmt.Printf("Service is %s from %s\n\n", stateString, serviceimpl.DefaultServiceInstallFilePath())

	serviceimpl.ServiceHelp()

	return nil
}

type StartServiceCmd struct{}

func (r *StartServiceCmd) Run(gOpts *options.GlobalOptions) error {
	if !serviceimpl.Admin() {
		return fmt.Errorf("service related actions require Admin (%s)", serviceimpl.HelpAdmin())
	}
	s, err := common()
	if err != nil {
		return err
	}
	serviceimpl.ServiceHelp()

	return service.Control(s, "start")
}

type StopServiceCmd struct{}

func (r *StopServiceCmd) Run(gOpts *options.GlobalOptions) error {
	if !serviceimpl.Admin() {
		return fmt.Errorf("service related actions require Admin (%s)", serviceimpl.HelpAdmin())
	}
	s, err := common()
	if err != nil {
		return err
	}
	return service.Control(s, "stop")
}

type RestartServiceCmd struct{}

func (r *RestartServiceCmd) Run(gOpts *options.GlobalOptions) error {
	if !serviceimpl.Admin() {
		return fmt.Errorf("service related actions require Admin (%s)", serviceimpl.HelpAdmin())
	}
	s, err := common()
	if err != nil {
		return err
	}

	// TODO: lib does support restart, but it causes hang that eats
	// up full core CPU fairly regularly on OSX.
	// see https://github.com/workbenchapp/worknet/issues/212
	if err := service.Control(s, "stop"); err != nil {
		return err
	}
	return service.Control(s, "start")
}

type InstallServiceCmd struct{}

func (r *InstallServiceCmd) Run(gOpts *options.GlobalOptions) error {
	if !serviceimpl.Admin() {
		return fmt.Errorf("service related actions require Admin (%s)", serviceimpl.HelpAdmin())
	}
	// check if the service is already installed

	s, err := common()
	if err != nil {
		return fmt.Errorf("service may already be installed: %s", err)
	}
	status, err := s.Status()
	if err != nil {
		// can get an error "not installed", "installed, but not run yet"
	} else {
		if status == service.StatusStopped {
			return fmt.Errorf("service is already installed, but stopped")
		}
		if status == service.StatusRunning {
			return fmt.Errorf("service is already installed, and is running")
		}
	}

	// if the running program isn't in the right place, copy it to there
	ePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, err := filepath.EvalSymlinks(ePath)
	if err != nil {
		return err
	}
	if exePath != serviceimpl.DefaultServiceInstallFilePath() {
		err := os.MkdirAll(serviceimpl.DefaultServiceInstallDir, 0755)
		if err != nil {
			return err
		}
		fmt.Printf("copying binary from %s to %s\n", exePath, serviceimpl.DefaultServiceInstallFilePath())
		_, err = copy(exePath, serviceimpl.DefaultServiceInstallFilePath())
		if err != nil {
			return err
		}
	}

	// TODO: imo we should also start it.

	return service.Control(s, "install")
}

type UninstallServiceCmd struct{}

func (r *UninstallServiceCmd) Run(gOpts *options.GlobalOptions) error {
	if !serviceimpl.Admin() {
		return fmt.Errorf("service related actions require Admin (%s)", serviceimpl.HelpAdmin())
	}
	s, err := common()
	if err != nil {
		return err
	}
	status, err := s.Status()
	if err == nil {
		if status == service.StatusUnknown {
			return fmt.Errorf("service does not appear to the installed")
		}
		if status == service.StatusRunning {
			return fmt.Errorf("please stop the service first")
		}
	}

	return service.Control(s, "uninstall")
}

var logger service.Logger

func common() (service.Service, error) {
	options := make(service.KeyValue)
	//OSX
	options["KeepAlive"] = true
	options["RunAtLoad"] = true
	// POSIX
	options["LogOutput"] = true
	options["Restart"] = "always"
	// Windows
	options["DelayedAutoStart"] = true
	options["StartType"] = "automatic"
	svcConfig := &service.Config{
		Name:        "daonetes",
		DisplayName: "DAOnetes-agent service",
		Description: "DAOnetes agent.",

		Executable: serviceimpl.DefaultServiceInstallFilePath(),
		Arguments:  []string{"agent", "--feature-flags=dns"},
		Option:     options,
	}

	prg := &systemProgram{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating new service object: %s", err)
	}
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatalf("OOPS: %s", err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()
	return s, err
}

///////////////////////////////////////////
// https://pkg.go.dev/github.com/kardianos/service
type systemProgram struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *systemProgram) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	p.ctx, p.cancel = context.WithCancel(context.Background())
	go p.run()

	return nil
}
func (p *systemProgram) run() {
	// setup Logging (chose this to setup for https://pkg.go.dev/go.opentelemetry.io/otel#SetLogger)
	zc := zap.NewDevelopmentConfig()
	zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zc.Level = zap.NewAtomicLevelAt(zapcore.Level(-cli.Verbose))
	z, err := zc.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialise logging: %v\n", err))
	}
	// Do work here
	cmd := DaoletCmd{}
	gOpts := options.GlobalOptions{ // TODO: this must die - logr can put the log into the context..
		Ctx: p.ctx,
		Log: zapr.NewLogger(z), // TODO: no, this needs to use service.Logger somehow
	}
	cmd.Run(&gOpts)
}
func (p *systemProgram) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	p.cancel()
	return nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	destination.Chmod(0755)
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
