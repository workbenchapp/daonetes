package solana

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gagliardetto/solana-go"
	"github.com/mitchellh/mapstructure"
	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/pkg/pointer"
	"github.com/portto/solana-go-sdk/program/assotokenprog"
	"github.com/portto/solana-go-sdk/program/metaplex/tokenmeta"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/program/tokenprog"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
)

/**

So this is going to be interesting.

I want there to be a SPEC hub, where there can be Global, Dao, and User 'SPEC tokens' (mmm, what about daolet?), all of which can be sent to one, or more daolets

>> its also possible that a doalet make a 'SPEC token' - basically like a user doing 'docker run' on a system, and then wanting to share with someone else.

This means that its _extremely_ important that the daolet code verify that the token comes from a trusted account, where trust can only be determined from on-chain info

Global tokens would be created (and thus paid for) by us, but clearly, i need to switch the fee payer for transfering the token to the token receiver

**/

const TOKENSYMBOL = "SPEX" //"FIGY"

func MakeNewToken(gOpts *options.GlobalOptions, receiver, feePayer *types.Account, specName string, specURL url.URL) {
	c := client.NewClient(program.GetClusterByName("").RPC)

	mint := types.NewAccount()
	fmt.Printf("NFT: %v\n", mint.PublicKey.ToBase58())

	collection := types.NewAccount()
	fmt.Println(base58.Encode(collection.PrivateKey))
	fmt.Printf("collection: %v\n", collection.PublicKey.ToBase58())

	ata, _, err := common.FindAssociatedTokenAddress(receiver.PublicKey, mint.PublicKey)
	if err != nil {
		log.Fatalf("failed to find a valid ata, err: %v", err)
	}

	tokenMetadataPubkey, err := tokenmeta.GetTokenMetaPubkey(mint.PublicKey)
	if err != nil {
		log.Fatalf("failed to find a valid token metadata, err: %v", err)

	}
	tokenMasterEditionPubkey, err := tokenmeta.GetMasterEdition(mint.PublicKey)
	if err != nil {
		log.Fatalf("failed to find a valid master edition, err: %v", err)
	}

	mintAccountRent, err := c.GetMinimumBalanceForRentExemption(gOpts.Ctx, tokenprog.MintAccountSize)
	if err != nil {
		log.Fatalf("failed to get mint account rent, err: %v", err)
	}

	recentBlockhashResponse, err := c.GetRecentBlockhash(gOpts.Ctx)
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{mint, *feePayer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feePayer.PublicKey,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				sysprog.CreateAccount(sysprog.CreateAccountParam{
					From:     feePayer.PublicKey,
					New:      mint.PublicKey,
					Owner:    common.TokenProgramID,
					Lamports: mintAccountRent,
					Space:    tokenprog.MintAccountSize,
				}),
				tokenprog.InitializeMint(tokenprog.InitializeMintParam{
					Decimals: 0,
					Mint:     mint.PublicKey,
					MintAuth: feePayer.PublicKey,
				}),
				tokenmeta.CreateMetadataAccountV2(tokenmeta.CreateMetadataAccountV2Param{
					Metadata:                tokenMetadataPubkey,
					Mint:                    mint.PublicKey,
					MintAuthority:           feePayer.PublicKey,
					Payer:                   feePayer.PublicKey,
					UpdateAuthority:         feePayer.PublicKey,
					UpdateAuthorityIsSigner: true,
					IsMutable:               true,
					Data: tokenmeta.DataV2{
						Name:                 specName,
						Symbol:               TOKENSYMBOL,
						Uri:                  specURL.String(),
						SellerFeeBasisPoints: 100,
						Creators: &[]tokenmeta.Creator{
							{
								Address:  feePayer.PublicKey,
								Verified: true,
								Share:    100,
							},
						},
						Collection: &tokenmeta.Collection{
							Verified: false,
							Key:      collection.PublicKey,
						},
						Uses: &tokenmeta.Uses{
							UseMethod: tokenmeta.Burn,
							Remaining: 10,
							Total:     10,
						},
					},
				}),
				assotokenprog.CreateAssociatedTokenAccount(assotokenprog.CreateAssociatedTokenAccountParam{
					Funder:                 feePayer.PublicKey,
					Owner:                  receiver.PublicKey,
					Mint:                   mint.PublicKey,
					AssociatedTokenAccount: ata,
				}),
				tokenprog.MintTo(tokenprog.MintToParam{
					Mint:   mint.PublicKey,
					To:     ata,
					Auth:   feePayer.PublicKey,
					Amount: 1,
				}),
				tokenmeta.CreateMasterEditionV3(tokenmeta.CreateMasterEditionParam{
					Edition:         tokenMasterEditionPubkey,
					Mint:            mint.PublicKey,
					UpdateAuthority: feePayer.PublicKey,
					MintAuthority:   feePayer.PublicKey,
					Metadata:        tokenMetadataPubkey,
					Payer:           feePayer.PublicKey,
					MaxSupply:       pointer.Uint64(0),
				}),
			},
		}),
	})
	if err != nil {
		log.Fatalf("failed to new a tx, err: %v", err)
	}

	// Send&Confirm...
	sig, err := c.SendTransaction(gOpts.Ctx, tx)
	if err != nil {
		log.Fatalf("failed to send tx, err: %v", err)
	}

	fmt.Printf("SUCCESS: created new SPEC-token for %s: sig: %s\n", receiver.PublicKey, sig)
}

type SPLTokenDataStruct struct {
	Program string
	Space   float64
	Parsed  struct {
		Type string
		Info struct {
			IsNative    bool
			Mint        string
			Owner       string
			Stake       string
			TokenAmount struct {
				Amount         string
				Decimals       float64
				UiAmount       float64
				UiAmountString string
			}
		}
	}
}

type TokenInfo struct {
	TokenAccountPubKey string
	TokenCount         float64
	Meta               *tokenmeta.Metadata
}

// Also airdropped myself some USDC token - https://onmyway133.com/posts/how-to-check-spl-token-balance-on-solana/
func GetSpecTokenList(gOpts *options.GlobalOptions, account *common.PublicKey) (tokens []*TokenInfo, err error) {
	c := client.NewClient(program.GetClusterByName("").RPC)

	// https://docs.solana.com/developing/clients/jsonrpc-api#gettokenaccountsbyowner
	res, err := c.RpcClient.GetTokenAccountsByOwnerWithConfig(
		gOpts.Ctx,
		account.ToBase58(),
		rpc.GetTokenAccountsByOwnerConfigFilter{
			// can filter by mint or programid (the latter...)
			ProgramId: common.TokenProgramID.String(),
		}, rpc.GetTokenAccountsByOwnerConfig{
			Encoding: rpc.GetTokenAccountsByOwnerConfigEncodingJsonParsed,
		})
	if err != nil {
		return tokens, fmt.Errorf("failed to get accounts, err: %v", err)
	}
	if len(res.Result.Value) == 0 {
		return tokens, nil
	}
	for _, a := range res.Result.Value {
		//tokenData := (a.Account.Data).(SPLTokenDataStruct)
		var tokenData SPLTokenDataStruct
		err := mapstructure.Decode(a.Account.Data, &tokenData)
		if err != nil {
			gOpts.VDebug().Info("failed to convert data map to SPLTokenDataStruct", "error", err)
			continue
		}
		gOpts.Debug().Info("TokenAccount",
			"PubKey", a.Pubkey,
			"Stake", float64(a.Account.Lamports)/float64(solana.LAMPORTS_PER_SOL),
			"TokenCount", tokenData.Parsed.Info.TokenAmount.UiAmount,
		)
		gOpts.VDebug().Info("TokenAccount", "obj", tokenData.Parsed)

		mint, err := c.RpcClient.GetAccountInfoWithConfig(gOpts.Ctx,
			tokenData.Parsed.Info.Mint,
			rpc.GetAccountInfoConfig{
				Encoding: rpc.GetAccountInfoConfigEncodingJsonParsed,
			})
		if err != nil {
			gOpts.VDebug().Info("failed to get mint", "error", err)
			continue
		}
		gOpts.Debug().Info("Mint",
			"PubKey", tokenData.Parsed.Info.Mint,
			"Stake", float64(mint.Result.Value.Lamports)/float64(solana.LAMPORTS_PER_SOL),
		)
		gOpts.VDebug().Info("mint", "obj", mint)

		metaDataAccount, err := tokenmeta.GetTokenMetaPubkey(common.PublicKeyFromString(tokenData.Parsed.Info.Mint))
		if err != nil {
			gOpts.VDebug().Info("failed to get mint metadata", "error", err)
			continue
		}
		meta, err := c.GetAccountInfo(gOpts.Ctx,
			metaDataAccount.String(),
		)
		if err != nil {
			gOpts.VDebug().Info("failed to get mint metadata info", "error", err)
		}
		gOpts.Debug().Info("Meta",
			"PubKey", metaDataAccount.String(),
			"Stake", float64(meta.Lamports)/float64(solana.LAMPORTS_PER_SOL),
		)
		metadata, err := tokenmeta.MetadataDeserialize(meta.Data)
		if err != nil {
			gOpts.Debug().Info(
				"failed to parse mint metadata info (likely not a metaplex format)",
				"error", err,
				"meta", meta,
			)
			continue
		}
		gOpts.VDebug().Info("metadata", "obj", metadata)
		var m tokenmeta.Metadata
		err = mapstructure.Decode(metadata, &m)
		if err != nil {
			gOpts.Log.Error(err, "failed to convert meta map to tokenmeta.Metadata")
			continue
		}
		gOpts.VDebug().Info("metadata", "decoded", m)

		tokens = append(tokens, &TokenInfo{
			TokenAccountPubKey: a.Pubkey,
			TokenCount:         tokenData.Parsed.Info.TokenAmount.UiAmount,
			Meta:               &m,
		})

	}
	return tokens, nil
}
