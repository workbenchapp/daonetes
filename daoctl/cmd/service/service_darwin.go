//go:build darwin
// +build darwin

package service

import (
	"fmt"
	"path/filepath"
)

// TODO: actually, the app should be in somethign more like
// /Applications/daonetes.app/Contents/Resources/bin/

const DefaultServiceInstallDir = "/usr/local/bin/"
const defaultServiceFileName = "daoctl"

func DefaultServiceInstallFilePath() string {
	return filepath.Join(DefaultServiceInstallDir, defaultServiceFileName)
}

func ServiceHelp() {
	fmt.Printf("Run 'launchctl status daonetes' to get status\n")
	//sudo tail -f /usr/local/var/log/daonetes.*
}
