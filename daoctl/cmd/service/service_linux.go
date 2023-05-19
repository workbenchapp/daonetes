//go:build linux
// +build linux

package service

import (
	"fmt"
	"path/filepath"
)

const DefaultServiceInstallDir = "/usr/local/bin/"
const defaultServiceFileName = "daoctl"

func DefaultServiceInstallFilePath() string {
	return filepath.Join(DefaultServiceInstallDir, defaultServiceFileName)
}

func ServiceHelp() {
	fmt.Printf("Run 'sudo systemctl status daonetes' to get status\n")
	fmt.Printf("Run 'sudo journalctl -fu daonetes' to get status\n")
}
