package cmd

import (
	"fmt"

	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/version"
)

type VersionCmd struct {
}

func (r *VersionCmd) Run(gOpts *options.GlobalOptions) error {
	fmt.Printf("Version: %s\n", version.GetVersionString())
	fmt.Printf("Revision: %s\n", version.GetBuildRevision())
	fmt.Printf("Build date: %s\n", version.GetBuildDate()) // `go run .`` returns an empty string..

	return nil
}
