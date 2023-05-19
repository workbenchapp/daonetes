package orchestrate

import (
	"github.com/portto/solana-go-sdk/common"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
)

func Orchestrate(gOpts *options.GlobalOptions, publickey common.PublicKey, tokens []*solana.TokenInfo) error {
	for _, token := range tokens {
		// TODO: test that the token was put there by a trusted account
		// TODO: need a RmCompose if count <=0
		if token.TokenCount > 0.0 {
			RunCompose(gOpts, publickey, token)
		}
	}

	return nil
}
