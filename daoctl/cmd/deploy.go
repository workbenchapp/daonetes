package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	bin "github.com/gagliardetto/binary"
	gagliardetto "github.com/gagliardetto/solana-go"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/smartwalletutils"
)

type DeployCmd struct {
	List     DeployCmdList     `cmd:"" help:"List deployments" default:"1"`
	Create   DeployCmdCreate   `cmd:"" help:"Create a new deployment"`
	Schedule DeployCmdSchedule `cmd:"" help:"Send deploy tokens to a device"`
}

type DeployCmdList struct{}

type DeployCmdCreate struct {
	Name     string `arg:"" help:"Name of the deployment"`
	SpecPDA  string `arg:"" help:"PDA for spec to deploy"`
	Replicas uint8  `arg:"replicas" default:"1" help:"Number of replicas (tokens) to create"`
}

type DeployCmdSchedule struct {
	Name       string `arg:"" help:"Name of the deployment to schedule"`
	NodePubKey string `arg:""`
	Replicas   uint8  `arg:"" default:"1"`
}

type DeploymentPDAs struct {
	WorkGroup        gagliardetto.PublicKey
	Deployment       gagliardetto.PublicKey
	DeploymentMint   gagliardetto.PublicKey
	DeploymentTokens gagliardetto.PublicKey
}

func NewDeploymentPDAs(
	walletPubKey gagliardetto.PublicKey, deploymentName string,
) (*DeploymentPDAs, error) {
	groupPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
		walletPubKey.Bytes(),
		[]byte("work_group"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return nil, fmt.Errorf("couldn't find PDA: %s", err)
	}

	deploymentPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
		walletPubKey.Bytes(),
		[]byte(deploymentName),
		[]byte("deployment"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return nil, fmt.Errorf("couldn't find deployment PDA: %s", err)
	}

	deploymentMintPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
		deploymentPDA.Bytes(),
		[]byte("deployment_mint"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return nil, fmt.Errorf("couldn't find deployment mint PDA: %s", err)
	}

	deploymentTokensPDA, _, err := gagliardetto.FindProgramAddress([][]byte{
		deploymentPDA.Bytes(),
		[]byte("deployment_tokens"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return nil, fmt.Errorf("couldn't find deployment token account PDA: %s", err)
	}

	return &DeploymentPDAs{
		WorkGroup:        groupPDA,
		Deployment:       deploymentPDA,
		DeploymentMint:   deploymentMintPDA,
		DeploymentTokens: deploymentTokensPDA,
	}, nil
}

func (r *DeployCmdList) Run(gOpts *options.GlobalOptions) error {
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
	fmt.Fprintln(tw, "NAME\tSPEC KEY\tKEY\tREPLICAS")

	defer func() {
		tw.Flush()
	}()

	if len(group.Deployments) == 0 {
		return nil
	}

	deployAccountsResp, err := sender.Client.GetMultipleAccounts(ctx, group.Deployments...)
	if err != nil {
		return fmt.Errorf("failed getting spec accounts: %s", err)
	}

	for i, deployAccount := range deployAccountsResp.Value {
		specDecoder := bin.NewDecoderWithEncoding(deployAccount.Data.GetBinary(), bin.EncodingBorsh)
		deployment := worknet.Deployment{}
		if err := deployment.UnmarshalWithDecoder(specDecoder); err != nil {
			return fmt.Errorf("decoding deployment failed: %s", err)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", deployment.Name, deployment.Spec.String(), group.Deployments[i].String(), deployment.Replicas)
	}

	return nil
}

func (r *DeployCmdCreate) Run(gOpts *options.GlobalOptions) error {
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

	specKey := gagliardetto.MustPublicKeyFromBase58(r.SpecPDA)

	deployPDA, err := NewDeploymentPDAs(pdas.DerivedWallet.Key, r.Name)
	if err != nil {
		return err
	}

	deployInst, err := worknet.NewCreateDeploymentInstruction(
		r.Name,
		r.Replicas,
		pdas.DerivedWallet.Key,
		deployPDA.Deployment,
		specKey,
		deployPDA.DeploymentMint,
		deployPDA.DeploymentTokens,
		deployPDA.WorkGroup,
		gagliardetto.SystemProgramID,
		gagliardetto.TokenProgramID,
		gagliardetto.SysVarRentPubkey,
	).ValidateAndBuild()
	if err != nil {
		return err
	}

	insts, err := smartwalletutils.WrapTransactions(
		ctx,
		sender.Client,
		*walletPubKey,
		pdas,
		[]gagliardetto.Instruction{deployInst},
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

	return nil
}

func (r *DeployCmdSchedule) Run(gOpts *options.GlobalOptions) error {
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

	deployPDA, err := NewDeploymentPDAs(pdas.DerivedWallet.Key, r.Name)
	if err != nil {
		return err
	}

	deviceAuthority := gagliardetto.MustPublicKeyFromBase58(r.NodePubKey)

	devicePDA, _, err := gagliardetto.FindProgramAddress([][]byte{
		deviceAuthority.Bytes(),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return fmt.Errorf("couldn't find device authority PDA: %s", err)
	}

	deviceTokens, _, err := gagliardetto.FindProgramAddress([][]byte{
		deviceAuthority.Bytes(),
		deployPDA.Deployment.Bytes(),
		[]byte("device_tokens"),
	}, program.WORKNET_V1_PROGRAM_PUBKEY)
	if err != nil {
		return fmt.Errorf("couldn't find device authority PDA: %s", err)
	}

	scheduleInst, err := worknet.NewScheduleInstruction(
		r.Replicas,
		pdas.DerivedWallet.Key,
		deployPDA.WorkGroup,
		deployPDA.Deployment,
		deployPDA.DeploymentMint,
		deployPDA.DeploymentTokens,
		devicePDA,
		deviceAuthority,
		deviceTokens,
		gagliardetto.SystemProgramID,
		gagliardetto.TokenProgramID,
		gagliardetto.SysVarRentPubkey,
	).ValidateAndBuild()
	if err != nil {
		return err
	}

	insts, err := smartwalletutils.WrapTransactions(
		ctx,
		sender.Client,
		*walletPubKey,
		pdas,
		[]gagliardetto.Instruction{scheduleInst},
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

	return nil
}
