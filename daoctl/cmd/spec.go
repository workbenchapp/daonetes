package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	bin "github.com/gagliardetto/binary"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/smartwalletutils"
)

const MAX_SPEC_URL_LEN = 64

type SpecCmdListCmd struct {
}

type SpecCmdRegisterCmd struct {
	Name string `help:"SPEC Name" arg:""`
	File string `help:"File for SPEC. Formats accepted: docker-compose.yml" required:"" short:"f"`
}

type SpecCmd struct {
	List     SpecCmdListCmd     `cmd:"" default:"1" help:"List SPECs"`
	Register SpecCmdRegisterCmd `cmd:"" help:"Register a SPEC"`
	// Stop...
}

func (r *SpecCmdListCmd) Run(gOpts *options.GlobalOptions) error {
	ctx := gOpts.Ctx

	pdas, err := smartwalletutils.SmartWalletAndGroupPDAs(ctx, nil)
	if err != nil {
		return err
	}

	group, _, err := solana.WorkGroupFromPubKey(ctx, pdas.DerivedWallet.Key)
	if err != nil {
		return fmt.Errorf("couldn't get workgroup from pubkey (%s): %s", pdas.DerivedWallet.Key.String(), err)
	}

	sender, err := solana.NewTransactionSender(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create transaction sender: %s", err)
	}

	// settings mostly cribbed from docker
	tw := tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tCONTAINERS\tKEY")

	defer func() {
		tw.Flush()
	}()

	if len(group.Specs) == 0 {
		return nil
	}

	specAccountsResp, err := sender.Client.GetMultipleAccounts(ctx, group.Specs...)
	if err != nil {
		return fmt.Errorf("failed getting spec accounts: %s", err)
	}

	for i, specAccount := range specAccountsResp.Value {
		specDecoder := bin.NewDecoderWithEncoding(specAccount.Data.GetBinary(), bin.EncodingBorsh)
		spec := worknet.WorkSpec{}
		if err := spec.UnmarshalWithDecoder(specDecoder); err != nil {
			return fmt.Errorf("decoding spec failed: %s", err)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", spec.Name, spec.UrlOrContents[0:MAX_SPEC_URL_LEN], group.Specs[i].String())
	}

	return nil
}

func (r *SpecCmdRegisterCmd) Run(gOpts *options.GlobalOptions) error {
	// TODO: Fix after breaking spec changes
	/*
		ctx := gOpts.Ctx

		walletPrivKey, walletPubKey, err := solana.MustGetWallet(ctx, gOpts)
		if err != nil {
			return fmt.Errorf("failed to get your wallet: %s", err)
		}

		pdas, err := smartwalletutils.SmartWalletAndGroupPDAs(ctx, nil)
		if err != nil {
			return err
		}

		sender, err := solana.NewTransactionSender(ctx)
		if err != nil {
			return fmt.Errorf("couldn't create transaction sender: %s", err)
		}

		composeYAMLContent, err := ioutil.ReadFile(r.File)
		if err != nil {
			return fmt.Errorf("loading compose file failed: %s", err)
		}

		specPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
			pdas.DerivedWallet.Key.Bytes(),
			[]byte(r.Name),
			[]byte("spec"),
		}, program.WORKNET_V1_PROGRAM_PUBKEY)
		if err != nil {
			return fmt.Errorf("couldn't find PDA: %s", err)
		}

			createSpecInst, err := worknet.NewCreateWorkSpecInstruction(
				r.Name,
				specContainers,
				pdas.DerivedWallet.Key,
				specPDA,
				pdas.Group.Key,
				gagliardetto.SystemProgramID,
			).ValidateAndBuild()
			if err != nil {
				return fmt.Errorf("could not validate CreateWorkSpec instruction: %s", err)
			}

			insts, err := smartwalletutils.WrapTransactions(
				ctx,
				sender.Client,
				*walletPubKey,
				pdas,
				[]gagliardetto.Instruction{createSpecInst},
			)
			if err != nil {
				return err
			}

			_, err = sender.SendAndConfirmTransaction(
				ctx,
				insts,
				solana.SignerKeys{*walletPubKey: walletPrivKey},
			)
			if err != nil {
				return fmt.Errorf("sending and confirming transaction failed: %s", err)
			}
	*/

	return nil
}
