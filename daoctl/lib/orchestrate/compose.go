package orchestrate

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/portto/solana-go-sdk/common"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/util"
)

func RunCompose(gOpts *options.GlobalOptions, daoletPublicKey common.PublicKey, token *solana.TokenInfo) error {
	resp, err := http.Get(token.Meta.Data.Uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// save the yaml file to a dir based on <daolet>/<token>/docker-compose.yml so we can deal with more than one daolet running
	// this may be a mistake, and it might be better to not separate by daolet - but what if the user has iSCSI to their home dir, and has multiple diskless workstations
	daoletCfgDir, err := gOpts.GetConfigDir("WorkNet")
	if err != nil {
		return err
	}
	composeDir := filepath.Join(daoletCfgDir, daoletPublicKey.String(), token.TokenAccountPubKey)
	err = os.MkdirAll(composeDir, os.ModeDir|0700)
	if err != nil {
		return err
	}
	composeFile := filepath.Join(composeDir, "docker-compose.yml")

	gOpts.Debug().Info("RunCompose", "composefile", composeFile)
	if _, err := os.Stat(composeFile); !os.IsNotExist(err) {
		// TODO: test if there are any changes...
		gOpts.Debug().Info("skipping RunCompose, file already exists", "composefile", composeFile)

		return nil
	}

	outFile, err := os.Create(composeFile)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return err
	}

	err = util.RunCmd(gOpts, composeDir, "docker", "compose", "up", "-d", "--wait")
	if err != nil {
		return err
	}

	return nil
}
