package util

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/workbenchapp/worknet/daoctl/lib/options"
)

// TODO: move these to a logical place, don't make bags of globals
// TODO: move to somewhere sane
func RunCmd(gOpts *options.GlobalOptions, runInDir, command string, args ...string) error {
	logCmd := fmt.Sprintf("%s %v", command, args)
	gOpts.Debug().Info("Run", "cmd", logCmd)

	cmd := exec.Command(command, args...)
	cmd.Dir = runInDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		gOpts.VDebug().Info("StdoutPipe", "error", err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		gOpts.VDebug().Info("StderrPipe", "error", err)
		return err
	}
	defer func() {
		_ = stdout.Close()
		_ = stderr.Close()
	}()

	err = cmd.Start()
	if err != nil {
		gOpts.VDebug().Info("cmd.Start", "error", err)
		return err
	}

	errscanner := bufio.NewScanner(stderr)
	go func() {
		for errscanner.Scan() {
			gOpts.Log.Info(errscanner.Text())
		}
	}()
	outscanner := bufio.NewScanner(stdout)
	go func() {
		for outscanner.Scan() {
			gOpts.Log.Info(outscanner.Text())
		}
	}()
	if err := cmd.Wait(); err != nil {
		gOpts.Log.Error(err, "cmd.Wait")
	}
	return err
}
