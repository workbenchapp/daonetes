package memo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gagliardetto "github.com/gagliardetto/solana-go"
	"github.com/go-logr/logr"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/memoprog"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
)

// General purpose Key, value map data to any account.
// from https://github.com/workbenchapp/worknet/blob/638644869c0a4186efea7b383ddf8b27d97c25f8/daoctl/cmd/daolet.go

type DaoletInfoMemo = map[string]string

// TODO: add a GetFirstMemoWithKey(opts, pubkey, "peerkey")
// Gets the first Key:value info memo we parse successfully on pubkey
func GetFirstInfoMemo(ctx context.Context, pubkey common.PublicKey) (*DaoletInfoMemo, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(2).Info("Getting memo", "memoKey", pubkey.String())

	var daoletInfo DaoletInfoMemo
	memos, _ := getTransactionsMemos(ctx, &pubkey)
	for _, memo := range memos {
		log.V(2).Info("Got transaction memo", "memoKey", memo)
		// Decode it from JSON for now to simplify debugging, and encode it for efficiency later
		err := json.Unmarshal([]byte(memo), &daoletInfo)
		if err == nil {
			return &daoletInfo, nil
		}
	}
	return nil, nil
}

func getTransactionsMemos(ctx context.Context, account *common.PublicKey) (memos []string, err error) {
	// TODO: probably should not throw away all the tx metadata - date, who signed etc

	txs, err := GetTransactions(ctx, account.String())
	if err != nil {
		return memos, err
	}
	for _, tx := range txs {
		if tx.Memo != nil {
			m := *tx.Memo
			// remove the '[11] ' from the front of the string (its a length value)
			index := strings.Index(m, "] ")
			if index > -1 {
				m = m[index+1:]
			}

			memos = append(memos, m)
		}
	}
	return memos, nil
}

func GetEarliestTransaction(ctx context.Context, account common.PublicKey) (tx *rpc.GetSignaturesForAddressResult, err error) {
	log := logr.FromContextOrDiscard(ctx)
	txs, err := GetTransactions(ctx, account.String())
	if err != nil {
		return tx, err
	}
	if len(txs) == 0 {
		// This can happen if the validator network has had a mindwipe restart

		log.V(2).Info("hey whay? len zero", "account", account.String())
		return tx, nil
	}

	return &(txs[len(txs)-1]), nil
}

// TODO: this is horrible, and a clear sign that one can't rely on transactions as a source of info.
// workaround for "Note: Transactions processed before block 129912198 are not available at this time"
func GetTransactions(ctx context.Context, pubKey string) (txs []rpc.GetSignaturesForAddressResult, err error) {
	c := client.NewClient(program.GetClusterByName("").RPC)
	unfilteredTxs, err := c.GetSignaturesForAddress(ctx, pubKey)
	if err != nil {
		return txs, err
	}
	for _, tx := range unfilteredTxs {
		if tx.Slot > 129912198 {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func AddInfoMemo(ctx context.Context /*daoletWallet *types.Account, */, daoletInfo *DaoletInfoMemo) error {
	log := logr.FromContextOrDiscard(ctx)
	memoString, err := json.Marshal(daoletInfo)
	if err != nil {
		log.V(2).Info("Marshal", "value", daoletInfo, "err", err)
		return err
	}
	// TODO: AddMemoString fails - a stake account (what i originally did with daolets), can't spend SOL
	err = AddMemoString(ctx, string(memoString))
	if err != nil {
		log.V(2).Info("AddMemo", "value", string(memoString), "err", err)
	}
	return err
}

// TODO: this is just adding a memo to the device auth, not the device data
// TODO: might be interesting to see if the device auth can add a memo to the device data
func AddMemoString(ctx context.Context /*daoletWallet *types.Account, */, memo string) error {
	log := logr.FromContextOrDiscard(ctx)

	c := client.NewClient(program.GetClusterByName("").RPC)
	recentBlockhashResponse, err := c.GetRecentBlockhash(ctx)
	if err != nil {
		panic(err)
	}

	feePayerWallet, err := solana.MustGetAgentWallet(ctx)
	if err != nil {
		panic(err)
	}
	balance, err := solana.GetBalance(ctx, feePayerWallet)
	if err != nil {
		panic(err)
	}
	if balance < gagliardetto.LAMPORTS_PER_SOL { // TODO: yeah, random constant
		err = solana.Airdrop(ctx, feePayerWallet.PublicKey.String())
		if err != nil {
			panic(err)
		}
		for i := 0; i < 4; i++ {
			time.Sleep(777 * time.Millisecond)
			balance, _ = solana.GetBalance(ctx, feePayerWallet)
			if balance > gagliardetto.LAMPORTS_PER_SOL {
				break
			}
		}
		// TODO: argh, need to wait til finalised...
	}
	log.Info("Checked Solana balance for fee payer",
		"feePayer", feePayerWallet,
		"solAmount", float64(balance)/float64(gagliardetto.LAMPORTS_PER_SOL),
	)

	// TODO: *types.Account should be replaced with types.Account..
	feePayer := types.Account{
		PublicKey:  feePayerWallet.PublicKey,
		PrivateKey: feePayerWallet.PrivateKey,
	}
	daoletWallet := feePayer
	daoUser := types.Account{
		PublicKey:  daoletWallet.PublicKey,
		PrivateKey: daoletWallet.PrivateKey,
	}
	//from := feePayer // get the user to pay the device registration fee

	// create a transfer tx
	// TODO: this only works for devnet, its a hack to allow DEMOv1 daolets to tell us onchain what their Peer keys are
	// TODO: note that this memo can't be trusted the way we are for DEMOv1 - any random account can do this...
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{feePayer, daoUser},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feePayerWallet.PublicKey,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				// The Memo here is something the daolet could use to set state info - like hostname/dns/publickey for other use
				memoprog.BuildMemo(memoprog.BuildMemoParam{
					SignerPubkeys: []common.PublicKey{daoUser.PublicKey},
					Memo:          []byte(memo),
				}),
			},
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to new a transaction, err: %v", err)
	}
	//spew.Dump(tx)

	// send tx
	sig, err := c.SendTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to send tx, err: %v", err)
	}

	log.V(2).Info("Sent memo transaction", "txnID", sig)
	return nil
}
