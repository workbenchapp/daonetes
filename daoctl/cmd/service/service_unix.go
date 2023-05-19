//go:build !windows
// +build !windows

package service

import "os"

func Admin() bool {
	return os.Geteuid() == 0
}
func HelpAdmin() string {
	return "rerun command using 'sudo'"
}
