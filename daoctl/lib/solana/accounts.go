package solana

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gagliardetto/solana-go"
	gagliardetto "github.com/gagliardetto/solana-go"
	"github.com/go-logr/logr"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/program"

	client2 "github.com/portto/solana-go-sdk/client"
	common2 "github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

func MustGetAgentWallet(ctx context.Context) (*types.Account, error) {
	agentConfig, err := options.Config()
	if err != nil {
		return nil, fmt.Errorf("error getting or creating agent config: %s", err)
	}

	activeNet, err := agentConfig.Active()
	if err != nil {
		return nil, err
	}

	ourWallet, err := MustGetAccount(ctx, "WorkNet", activeNet.KeyFile)
	if err != nil {
		return nil, err
	}

	return ourWallet, nil
}

// TODO: This should all be rolled into the gagliardetto bindings, but keeping
// here for now as to not break existing code.
func MustGetAccount(ctx context.Context, configName, keyFile string) (*types.Account, error) {
	log := logr.FromContextOrDiscard(ctx)
	daoletCfgDir, err := options.GetConfigDir(configName)
	if err != nil {
		return nil, err
	}
	ourWalletJSONFile := filepath.Join(daoletCfgDir, keyFile)

	var ourWallet *types.Account

	loadedAccount, err := GetUserAccount(ctx, ourWalletJSONFile)
	if err == nil {
		ourWallet = loadedAccount
		log.V(1).Info("Using existing public key", "ourWalletFile", ourWalletJSONFile, "ourWalletKey", ourWallet.PublicKey)
	}

	// TODO: cope with file exists, but we can't read it
	if ourWallet == nil {
		// presume it doesn't exist, and make a new one
		fmt.Printf("Creating agent wallet - writing to %s\n", ourWalletJSONFile)

		account := types.NewAccount()
		fmt.Println("Wallet public key:", account.PublicKey)

		err = os.MkdirAll(daoletCfgDir, os.ModeDir|0700)
		if err != nil {
			return nil, err
		}
		err = WritePrivateKeyAsJSON(ourWalletJSONFile, account.PrivateKey)
		if err != nil {
			return nil, err
		}
		ourWallet = &account

		// TODO: on initial creation, how do we say "i wanna store X bytes."
	}
	return ourWallet, nil
}

func MustGetDefaultWallet(gOpts *options.GlobalOptions) (*gagliardetto.PrivateKey, *gagliardetto.PublicKey, error) {
	daoletCfgDir, err := options.GetConfigDir("solana")
	if err != nil {
		return nil, nil, err
	}

	ourWalletJSONFile := path.Join(daoletCfgDir, "id.json")

	var (
		privKey gagliardetto.PrivateKey
		pubKey  gagliardetto.PublicKey
	)

	privKey, err = gagliardetto.PrivateKeyFromSolanaKeygenFile(ourWalletJSONFile)

	// TODO: This could be an error other than the file not existing.
	// We need to handle that case instead of assuming this error is
	// only the case we expect.
	if err == nil {
		pubKey = privKey.PublicKey()
		return &privKey, &pubKey, err
	}

	// presume it doesn't exist, and make a new one
	fmt.Printf("Creating agent wallet - writing to %s\n", ourWalletJSONFile)

	account := types.NewAccount()
	fmt.Println("Wallet public key:", account.PublicKey)

	err = WritePrivateKeyAsJSON(ourWalletJSONFile, account.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	privKey, err = gagliardetto.PrivateKeyFromSolanaKeygenFile(ourWalletJSONFile)
	if err == nil {
		pubKey = privKey.PublicKey()
	} else {
		return nil, nil, err
	}

	return &privKey, &pubKey, nil
}

func MustGetWallet(ctx context.Context, gOpts *options.GlobalOptions) (*gagliardetto.PrivateKey, *gagliardetto.PublicKey, error) {
	ourWalletJSONFile := ctx.Value(options.KeyFile).(string)
	if ourWalletJSONFile == "" {
		return MustGetDefaultWallet(gOpts)
	}

	var (
		privKey gagliardetto.PrivateKey
		pubKey  gagliardetto.PublicKey
	)

	privKey, err := gagliardetto.PrivateKeyFromSolanaKeygenFile(ourWalletJSONFile)
	if err != nil {
		return nil, nil, err
	}
	pubKey = privKey.PublicKey()
	return &privKey, &pubKey, nil
}

func GetUserAccount(ctx context.Context, walletJSONFile string) (*types.Account, error) {
	log := logr.FromContextOrDiscard(ctx)
	if walletJSONFile == "" {
		// share the User Key with the solana cli
		solanaCfgDir, err := options.GetConfigDir("solana")
		if err != nil {
			return nil, err
		}
		walletJSONFile = filepath.Join(solanaCfgDir, "id.json")
	}

	privateKey, err := gagliardetto.PrivateKeyFromSolanaKeygenFile(walletJSONFile)
	if err != nil {
		// TODO: what should we do if it doesn't exist
		//       I propose error out, and tell user to get their key, or add `--newuser`
		return nil, err
	}

	ourWallet := &types.Account{
		PublicKey:  common2.PublicKey(privateKey.PublicKey()),
		PrivateKey: ed25519.PrivateKey(privateKey),
	}
	log.V(1).Info("Existing User Account", "file", walletJSONFile, "PubKey", ourWallet.PublicKey)

	return ourWallet, nil
}

func GetAccountFromString(gOpts *options.GlobalOptions, seed string) (*types.Account, error) {
	privateKey, err := gagliardetto.PrivateKeyFromBase58(seed)
	if err != nil {
		return nil, err
	}
	ourWallet := &types.Account{
		PublicKey:  common2.PublicKey(privateKey.PublicKey()),
		PrivateKey: ed25519.PrivateKey(privateKey),
	}
	return ourWallet, nil
}

func GetBalance(ctx context.Context, ourWallet *types.Account) (uint64, error) {
	log := logr.FromContextOrDiscard(ctx)
	c := client2.NewClient(program.GetClusterByName("").RPC)
	balance, err := c.GetBalance(
		ctx,
		ourWallet.PublicKey.String(),
	)
	if err != nil {
		return 0, err
	}
	log.V(1).Info("Existing User Account",
		"PubKey", ourWallet.PublicKey,
		"Network", program.GetClusterByName("").Name,
		"balance", float64(balance)/float64(gagliardetto.LAMPORTS_PER_SOL),
	)

	return balance, nil
}

func GetRentExemption(gOpts *options.GlobalOptions) (uint64, error) {
	/*     <DATA_LENGTH_OR_MONIKER>    Length of data field in the account to calculate rent for, or moniker: [nonce,
	                                stake, system, vote]
	PS C:\Users\svend> solana -v rent stake
	RPC URL: https://api.mainnet-beta.solana.com
	Default Signer Path: C:\Users\svend\.config\solana\id.json
	Commitment: confirmed
	Rent per byte-year: 0.00000348 SOL
	Rent per epoch: 0.00000625 SOL
	Rent-exempt minimum: 0.00228288 SOL */

	dataSize := uint64(200) // magic constant for spl-stake at Apr2022

	c := client2.NewClient(program.GetClusterByName("").RPC)
	lamports, err := c.GetMinimumBalanceForRentExemption(gOpts.Ctx, dataSize)
	if err != nil {
		return 0, err
	}

	fmt.Printf("rent exemption (for %d bytes) needs %f SOL on %s\n", dataSize, float64(lamports)/float64(gagliardetto.LAMPORTS_PER_SOL), program.GetClusterByName("").Name)
	return lamports, nil
}

type KeyedAccount struct {
	Pubkey  string
	Account client2.AccountInfo
}

func Airdrop(ctx context.Context, publicKey string) error {
	log := logr.FromContextOrDiscard(ctx)
	c := client2.NewClient(program.GetClusterByName("").RPC)
	tx, err := c.RequestAirdrop(ctx, publicKey, solana.LAMPORTS_PER_SOL)
	log.V(1).Info("RequestAirdrop", "tx", tx, "err", err)

	// TODO: wait til we're funded?

	return err
}

func GetAccountInfo(gOpts *options.GlobalOptions, pubKey common2.PublicKey) (*KeyedAccount, error) {
	c := client2.NewClient(program.GetClusterByName("").RPC)
	res, err := c.GetAccountInfo(gOpts.Ctx, pubKey.String())
	if err != nil {
		return nil, err
	}
	account := KeyedAccount{
		Pubkey:  pubKey.String(),
		Account: res,
	}
	return &account, nil
}

// for new account creation and serialisation
type JSONableSlice []uint8

func (u JSONableSlice) MarshalJSON() ([]byte, error) {
	var result string
	if u == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", u)), ",")
	}
	return []byte(result), nil
}
func WritePrivateKeyAsJSON(jsonFilePath string, privateKey ed25519.PrivateKey) error {
	// convert to JSON
	var array JSONableSlice
	for _, v := range []uint8(privateKey) {
		array = append(array, v)
	}
	data, err := json.Marshal(array)
	if err != nil {
		return err
	}

	err = os.WriteFile(jsonFilePath, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
