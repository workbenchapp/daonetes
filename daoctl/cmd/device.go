package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	bin "github.com/gagliardetto/binary"
	gagliardettorpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/smartwalletutils"
)

type DeviceCmdListCmd struct {
}
type DeviceCmdRegisterCmd struct {
	DevicePubkey string `arg:"" help:"Public key for device's local keypair"`
}

type DeviceCmd struct {
	List     DeviceCmdListCmd     `cmd:"" default:"1" help:"List registered devices"`
	Register DeviceCmdRegisterCmd `cmd:"" help:"Register device"`
	// transfer device
}

func (r *DeviceCmdListCmd) Run(gOpts *options.GlobalOptions) error {
	ctx := gOpts.Ctx

	pdas, err := smartwalletutils.SmartWalletAndGroupPDAs(ctx, nil)
	if err != nil {
		return err
	}

	group, _, err := solana.WorkGroupFromPubKey(ctx, pdas.DerivedWallet.Key)
	if err != nil {
		return fmt.Errorf("couldn't get workgroup from pubkey (%s): %s", pdas.DerivedWallet.Key.String(), err)
	}

	// settings mostly cribbed from docker
	tw := tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0)
	fmt.Fprintln(tw, "HOSTNAME\tIPV4\tSTATUS\tAUTHORITY\tPDA")

	defer func() {
		tw.Flush()
	}()

	if len(group.Devices) == 0 {
		return nil
	}

	client := gagliardettorpc.New(options.SolanaCluster(ctx).RPC)

	deviceAccountResp, err := client.GetMultipleAccounts(ctx, group.Devices...)
	if err != nil {
		return fmt.Errorf("failed getting device accounts: %s", err)
	}

	for i, deviceAccount := range deviceAccountResp.Value {
		deviceDecoder := bin.NewDecoderWithEncoding(deviceAccount.Data.GetBinary(), bin.EncodingBorsh)
		device := worknet.Device{}
		if err := device.UnmarshalWithDecoder(deviceDecoder); err != nil {
			return fmt.Errorf("decoding device failed: %s", err)
		}
		fmt.Fprintf(tw,
			"%s\t%s\t%s\t%s\t%s\n",
			device.Hostname,
			fmt.Sprintf("%d.%d.%d.%d",
				device.Ipv4[0],
				device.Ipv4[1],
				device.Ipv4[2],
				device.Ipv4[3]),
			device.Status,
			device.DeviceAuthority.String(),
			group.Devices[i].String(),
		)
	}

	return nil
}

func (r *DeviceCmdRegisterCmd) Run(gOpts *options.GlobalOptions) error {
	// TODO: Fix from breaking license mint change
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

		devicePubkey, err := gagliardetto.PublicKeyFromBase58(r.DevicePubkey)
		if err != nil {
			return err
		}

		seeds := [][]byte{
			devicePubkey.Bytes(),
		}

		pdaPubkey, _, err := gagliardetto.FindProgramAddress(seeds, program.WORKNET_V1_PROGRAM_PUBKEY)
		if err != nil {
			return err
		}

		sender, err := solana.NewTransactionSender(ctx)
		if err != nil {
			return err
		}

		if _, err := sender.Client.GetAccountInfo(ctx, pdaPubkey); err != nil {
			if err != gagliardettorpc.ErrNotFound {
				return err
			}
		} else {
			return errors.New("device already registered")
		}

		registerDeviceInst, err := worknet.NewRegisterDeviceInstruction(
			devicePubkey,
			pdas.DerivedWallet.Key,
			pdaPubkey,
			options.LicenseMint(ctx),
			pdas.Group.Key,
			gagliardetto.SystemProgramID,
		).ValidateAndBuild()
		if err != nil {
			return err
		}

		// even 0 bytes account has a rent minimum, ~0.00000348 SOL (5/15/2022)
		rentMinimum, err := sender.Client.GetMinimumBalanceForRentExemption(ctx,
			0,
			gagliardettorpc.CommitmentFinalized,
		)
		if err != nil {
			return err
		}

		insts, err := smartwalletutils.WrapTransactions(
			ctx,
			sender.Client,
			*walletPubKey,
			pdas,
			[]gagliardetto.Instruction{
				registerDeviceInst,
				system.NewTransferInstruction(
					rentMinimum+5000,
					*walletPubKey,
					devicePubkey,
				).Build(),
			},
		)
		if err != nil {
			return err
		}

		if _, err := sender.SendAndConfirmTransaction(
			ctx,
			insts,
			solana.SignerKeys{*walletPubKey: walletPrivKey},
		); err != nil {
			return err
		}
	*/

	return nil
}
