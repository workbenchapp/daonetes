package solana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	gagliardetto "github.com/gagliardetto/solana-go"
	gagliardettorpc "github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	gagliardettorws "github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
)

type DaoletInfoMemo struct {
	// each daolet can add a memo with its hostname, public wg key, ip:port (for now)
	Hostname string
	PeerKey  string
	PublicIp string
}

type TransactionSender struct {
	Client   *gagliardettorpc.Client
	WSClient *gagliardettorws.Client
}

type SignerKeys map[gagliardetto.PublicKey]*gagliardetto.PrivateKey

func NewTransactionSender(ctx context.Context) (*TransactionSender, error) {
	cluster := options.SolanaCluster(ctx)

	wsClient, err := gagliardettorws.Connect(ctx, cluster.WS)
	if err != nil {
		return nil, err
	}

	sender := &TransactionSender{
		Client:   gagliardettorpc.New(cluster.RPC),
		WSClient: wsClient,
	}
	if err != nil {
		return sender, err
	}

	return sender, nil
}

func (sender *TransactionSender) SendAndConfirmTransaction(
	ctx context.Context,
	instructions []gagliardetto.Instruction,
	signers SignerKeys,
) (*gagliardetto.Signature, error) {
	blockHash, err := sender.Client.GetRecentBlockhash(ctx, gagliardettorpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	txn, err := gagliardetto.NewTransaction(
		instructions,
		blockHash.Value.Blockhash,
	)
	if err != nil {
		return nil, err
	}

	_, err = txn.Sign(func(key gagliardetto.PublicKey) *gagliardetto.PrivateKey {
		return signers[key]
	})
	if err != nil {
		return nil, err
	}

	spew.Dump(txn)

	sig, err := sendandconfirmtransaction.SendAndConfirmTransaction(
		ctx,
		sender.Client,
		sender.WSClient,
		txn,
	)

	return &sig, err
}

func WorkGroupFromPubKey(ctx context.Context, pubKey gagliardetto.PublicKey) (*worknet.WorkGroup, gagliardetto.PublicKey, error) {
	groupPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
		pubKey.Bytes(),
		[]byte("work_group"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return nil, groupPDA, fmt.Errorf("couldn't find PDA: %s", err)
	}

	client := gagliardettorpc.New(options.SolanaCluster(ctx).RPC)

	groupAccountResp, err := client.GetAccountInfo(ctx, groupPDA)
	if err != nil {
		if err != gagliardettorpc.ErrNotFound {
			return nil, groupPDA, errors.New("no group account found, create one with 'daoctl group init'")
		}
		return nil, groupPDA, fmt.Errorf("error trying to get group account: %s", err)
	}

	group := worknet.WorkGroup{}
	groupAccount := groupAccountResp.Value
	groupDecoder := bin.NewDecoderWithEncoding(groupAccount.Data.GetBinary(), bin.EncodingBorsh)
	if err := group.UnmarshalWithDecoder(groupDecoder); err != nil {
		return nil, groupPDA, fmt.Errorf("couldn't unmarshal workgroup %s: %s", groupPDA.String(), err)
	}

	return &group, groupPDA, nil
}

func GetCurrentDaoletHostInfo(gOpts *options.GlobalOptions, pubkey common.PublicKey) (*DaoletInfoMemo, error) {
	var daoletInfo DaoletInfoMemo
	memos, _ := GetTransactionsMemos(gOpts, &pubkey)
	for _, memo := range memos {
		//fmt.Printf("         memo: %s\n", memo)
		// Decode it from JSON for now to simplify debugging, and encode it for efficiency later
		err := json.Unmarshal([]byte(memo), &daoletInfo)
		if err == nil {
			return &daoletInfo, nil
		}
	}
	return nil, nil
}

func GetTransactionsMemos(gOpts *options.GlobalOptions, account *common.PublicKey) (memos []string, err error) {
	txs, err := GetTransactions(gOpts, account.String())
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

func GetEarliestTransaction(gOpts *options.GlobalOptions, account common.PublicKey) (tx *rpc.GetSignaturesForAddressResult, err error) {
	txs, err := GetTransactions(gOpts, account.String())
	if err != nil {
		return tx, err
	}
	if len(txs) == 0 {
		// This can happen if the validator network has had a mindwipe restart

		gOpts.Log.Info("hey whay? len zero", "account", account.String())
		return tx, nil
	}

	return &(txs[len(txs)-1]), nil
}

// TODO: this is horrible, and a clear sign that one can't rely on transactions as a source of info.
// workaround for "Note: Transactions processed before block 129912198 are not available at this time"
func GetTransactions(gOpts *options.GlobalOptions, pubKey string) (txs []rpc.GetSignaturesForAddressResult, err error) {
	c := client.NewClient(program.GetClusterByName("").RPC)
	unfilteredTxs, err := c.GetSignaturesForAddress(gOpts.Ctx, pubKey)
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
