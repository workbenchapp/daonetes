package smartwalletutils

import (
	"context"
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
	gagliardetto "github.com/gagliardetto/solana-go"
	gagliardettorpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/smartwallet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
)

// gagliardetto and smartwallet both have their own types for account metadata.
// this function converts gagliardetto's type into smart wallet's for passing to
// the goki smart wallet program.
func gagliardettoAccountMetaToSmartWalletAccountMeta(
	accounts []*gagliardetto.AccountMeta,
	programID gagliardetto.PublicKey,
) []smartwallet.TXAccountMeta {
	txnAccountMetas := []smartwallet.TXAccountMeta{}
	for _, meta := range accounts {
		txnAccountMetas = append(txnAccountMetas, smartwallet.TXAccountMeta{
			Pubkey:     meta.PublicKey,
			IsSigner:   meta.IsSigner,
			IsWritable: meta.IsWritable,
		})
	}
	txnAccountMetas = append(txnAccountMetas, smartwallet.TXAccountMeta{
		Pubkey:     programID,
		IsSigner:   false,
		IsWritable: false,
	})
	return txnAccountMetas
}

func derivedWallet(
	ctx context.Context,
	smartWalletPDA gagliardetto.PublicKey,
) (gagliardetto.PublicKey, uint8, error) {
	// The smartWallet account itself can't receive SOL, because it has
	// data in it. Therefore we need a derived wallet to actually hold
	// the SOL and sign transactions.
	return gagliardetto.FindProgramAddress([][]byte{
		[]byte("GokiSmartWalletDerived"),
		smartWalletPDA.Bytes(),
		{0, 0, 0, 0, 0, 0, 0, 0}, // I assume this is index of the derived wallet? i.e. there could be multiple
	}, ctx.Value(options.GokiProgramPubkey).(gagliardetto.PublicKey))
}

type PDA struct {
	Key  gagliardetto.PublicKey
	Bump uint8
}

type SmartWalletAndGroupAccounts struct {
	SmartWallet   *PDA
	DerivedWallet *PDA
	Group         *PDA
}

func SmartWalletAndGroupPDAs(
	ctx context.Context,
	baseKey *gagliardetto.PublicKey,
) (*SmartWalletAndGroupAccounts, error) {
	smartWalletPDAStr := ctx.Value(options.SmartWalletAddress).(string)
	derivedWalletPDAStr := ctx.Value(options.DerivedWalletAddress).(string)

	var (
		smartWalletPDA    gagliardetto.PublicKey
		smartWalletBump   uint8
		derivedWalletPDA  gagliardetto.PublicKey
		derivedWalletBump uint8
		groupPDA          gagliardetto.PublicKey
		groupBump         uint8
		err               error
	)

	if smartWalletPDAStr == "" {
		smartWalletPDA, smartWalletBump, err = gagliardetto.FindProgramAddress([][]byte{
			[]byte("GokiSmartWallet"),
			baseKey.Bytes(),
		}, ctx.Value(options.GokiProgramPubkey).(gagliardetto.PublicKey))
		if err != nil {
			return nil, err
		}
	} else {
		smartWalletPDA = gagliardetto.MustPublicKeyFromBase58(smartWalletPDAStr)
	}

	fmt.Printf("smartWalletPDA: %s\n", smartWalletPDA.String())

	if derivedWalletPDAStr == "" {
		derivedWalletPDA, derivedWalletBump, err = derivedWallet(ctx, smartWalletPDA)
		if err != nil {
			return nil, err
		}
	} else {
		derivedWalletPDA = gagliardetto.MustPublicKeyFromBase58(smartWalletPDAStr)
	}

	fmt.Printf("derivedWalletPDA: %s\n", derivedWalletPDA.String())

	groupPDA, groupBump, err = gagliardetto.FindProgramAddress([][]byte{
		derivedWalletPDA.Bytes(),
		[]byte("work_group"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return nil, fmt.Errorf("couldn't find PDA: %s", err)
	}

	fmt.Printf("groupPDA: %s\n", groupPDA.String())

	return &SmartWalletAndGroupAccounts{
		SmartWallet: &PDA{
			Key:  smartWalletPDA,
			Bump: smartWalletBump,
		},
		DerivedWallet: &PDA{
			Key:  derivedWalletPDA,
			Bump: derivedWalletBump,
		},
		Group: &PDA{
			Key:  groupPDA,
			Bump: groupBump,
		},
	}, nil
}

func wrappedInstructionAccounts(
	walletPDA gagliardetto.PublicKey,
	wrappedInsts []gagliardetto.Instruction,
	derived bool,
) (gagliardetto.AccountMetaSlice, error) {
	appendedAccounts := map[gagliardetto.PublicKey]*gagliardetto.AccountMeta{}

	for _, wrappedInst := range wrappedInsts {
		for _, wrappedInstAccountMeta := range wrappedInst.Accounts() {
			// inner transaction has derived wallet as signer, but it is not
			// marked as signer when passed to the execute transaction, so don't
			// include that when creating the appended added account list
			if wrappedInstAccountMeta.PublicKey != walletPDA && derived {
				if acct, ok := appendedAccounts[wrappedInstAccountMeta.PublicKey]; ok {
					wrappedInstAccountMeta.IsWritable =
						wrappedInstAccountMeta.IsWritable || acct.IsWritable
				}
				appendedAccounts[wrappedInstAccountMeta.PublicKey] =
					wrappedInstAccountMeta
			}
		}
		appendedAccounts[wrappedInst.ProgramID()] = &gagliardetto.AccountMeta{
			IsWritable: false,
			IsSigner:   false,
			PublicKey:  wrappedInst.ProgramID(),
		}
	}

	appendedAccounts[walletPDA] = &gagliardetto.AccountMeta{
		IsWritable: true,
		IsSigner:   false,
		PublicKey:  walletPDA,
	}

	accountMetaSlice := gagliardetto.AccountMetaSlice{}

	for _, accountMeta := range appendedAccounts {
		accountMetaSlice = append(accountMetaSlice, accountMeta)
	}

	return accountMetaSlice, nil
}

func WrapTransactionsForGoki(
	ctx context.Context,
	client *gagliardettorpc.Client,
	walletPubKey gagliardetto.PublicKey,
	pdas *SmartWalletAndGroupAccounts,
	wrappedInsts []gagliardetto.Instruction,
) ([]gagliardetto.Instruction, error) {
	smartWalletPDA := pdas.SmartWallet.Key
	gokiWallet := smartwallet.SmartWallet{}
	numTxns := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	gokiWalletInfo, err := client.GetAccountInfo(ctx, smartWalletPDA)
	if err != nil {
		if err != gagliardettorpc.ErrNotFound {
			return nil, fmt.Errorf("error looking for smart wallet: %s", err)
		}
	} else {
		smartWalletDecoder := bin.NewDecoderWithEncoding(gokiWalletInfo.Value.Data.GetBinary(), bin.EncodingBorsh)
		if err := gokiWallet.UnmarshalWithDecoder(smartWalletDecoder); err != nil {
			return nil, fmt.Errorf("decoding smart wallet failed: %s", err)
		}

		binary.LittleEndian.PutUint64(numTxns, gokiWallet.NumTransactions)
	}

	gokiPubKey := ctx.Value(options.GokiProgramPubkey).(gagliardetto.PublicKey)

	txnPDA, txnPDABump, err := gagliardetto.FindProgramAddress([][]byte{
		[]byte("GokiTransaction"),
		smartWalletPDA.Bytes(),
		numTxns,
	}, gokiPubKey)
	if err != nil {
		return nil, err
	}

	smartWalletTxnInstructions := []smartwallet.TXInstruction{}

	for _, inst := range wrappedInsts {
		data, err := inst.Data()
		if err != nil {
			return nil, err
		}
		smartWalletTxnInstructions = append(smartWalletTxnInstructions, smartwallet.TXInstruction{
			ProgramId: inst.ProgramID(),
			Keys: gagliardettoAccountMetaToSmartWalletAccountMeta(
				inst.Accounts(),
				gokiPubKey,
			),
			Data: data,
		})
	}

	gokiCreateTxn, err := smartwallet.NewCreateTransactionInstruction(
		txnPDABump,
		smartWalletTxnInstructions,
		smartWalletPDA,
		txnPDA,
		walletPubKey,
		walletPubKey,
		gagliardetto.SystemProgramID,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	gokiApproveTxn, err := smartwallet.NewApproveInstruction(
		smartWalletPDA,
		txnPDA,
		walletPubKey,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	gokiExecuteTxnBuilder := smartwallet.NewExecuteTransactionInstruction(
		smartWalletPDA,
		txnPDA,
		walletPubKey,
	)

	extraAccounts, err := wrappedInstructionAccounts(smartWalletPDA, wrappedInsts, false)
	if err != nil {
		return nil, err
	}

	gokiExecuteTxnBuilder.AccountMetaSlice = append(
		gokiExecuteTxnBuilder.AccountMetaSlice, extraAccounts...,
	)

	gokiExecuteTxn, err := gokiExecuteTxnBuilder.ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	return []gagliardetto.Instruction{gokiCreateTxn, gokiApproveTxn, gokiExecuteTxn}, nil
}

func WrapTransactions(
	ctx context.Context,
	client *gagliardettorpc.Client,
	walletPubKey gagliardetto.PublicKey,
	pdas *SmartWalletAndGroupAccounts,
	wrappedInsts []gagliardetto.Instruction,
) ([]gagliardetto.Instruction, error) {
	smartWalletPDA := pdas.SmartWallet.Key
	derivedWalletPDA := pdas.DerivedWallet.Key
	derivedWalletBump := pdas.DerivedWallet.Bump
	gokiWallet := smartwallet.SmartWallet{}
	numTxns := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	gokiWalletInfo, err := client.GetAccountInfo(ctx, smartWalletPDA)
	if err != nil {
		if err != gagliardettorpc.ErrNotFound {
			return nil, fmt.Errorf("error looking for smart wallet: %s", err)
		}
	} else {
		smartWalletDecoder := bin.NewDecoderWithEncoding(gokiWalletInfo.Value.Data.GetBinary(), bin.EncodingBorsh)
		if err := gokiWallet.UnmarshalWithDecoder(smartWalletDecoder); err != nil {
			return nil, fmt.Errorf("decoding smart wallet failed: %s", err)
		}

		binary.LittleEndian.PutUint64(numTxns, gokiWallet.NumTransactions)
	}

	gokiPubKey := ctx.Value(options.GokiProgramPubkey).(gagliardetto.PublicKey)

	// b"GokiTransaction".as_ref(),
	// smart_wallet.key().to_bytes().as_ref(),
	// smart_wallet.num_transactions.to_le_bytes().as_ref()
	txnPDA, txnPDABump, err := gagliardetto.FindProgramAddress([][]byte{
		[]byte("GokiTransaction"),
		smartWalletPDA.Bytes(),
		numTxns,
	}, gokiPubKey)
	if err != nil {
		return nil, err
	}

	smartWalletTxnInstructions := []smartwallet.TXInstruction{}

	for _, inst := range wrappedInsts {
		data, err := inst.Data()
		if err != nil {
			return nil, err
		}
		smartWalletTxnInstructions = append(smartWalletTxnInstructions, smartwallet.TXInstruction{
			ProgramId: inst.ProgramID(),
			Keys: gagliardettoAccountMetaToSmartWalletAccountMeta(
				inst.Accounts(),
				program.WORKNET_V1_PROGRAM_PUBKEY,
			),
			Data: data,
		})
	}

	gokiCreateTxn, err := smartwallet.NewCreateTransactionInstruction(
		txnPDABump,
		smartWalletTxnInstructions,
		smartWalletPDA,
		txnPDA,
		walletPubKey,
		walletPubKey,
		gagliardetto.SystemProgramID,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	gokiApproveTxn, err := smartwallet.NewApproveInstruction(
		smartWalletPDA,
		txnPDA,
		walletPubKey,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	gokiExecuteTxnBuilder := smartwallet.NewExecuteTransactionDerivedInstruction(
		0, // derived wallet index
		derivedWalletBump,
		smartWalletPDA,
		txnPDA,
		walletPubKey,
	)

	extraAccounts, err := wrappedInstructionAccounts(
		derivedWalletPDA,
		wrappedInsts,
		true,
	)
	if err != nil {
		return nil, err
	}

	gokiExecuteTxnBuilder.AccountMetaSlice = append(
		gokiExecuteTxnBuilder.AccountMetaSlice, extraAccounts...,
	)

	gokiExecuteTxn, err := gokiExecuteTxnBuilder.ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	return []gagliardetto.Instruction{
		gokiCreateTxn,
		gokiApproveTxn,
		gokiExecuteTxn,
	}, nil
}
