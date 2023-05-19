package options

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	gagliardetto "github.com/gagliardetto/solana-go"
	gagliardettorpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/go-logr/logr"
	serviceimpl "github.com/workbenchapp/worknet/daoctl/cmd/service"
	"gopkg.in/yaml.v2"
)

type ContextKey string

const (
	URL                  ContextKey = "url"
	GokiProgramPubkey    ContextKey = "goki-program-pubkey"
	KeyFile              ContextKey = "key-file"
	SmartWalletAddress   ContextKey = "smart-wallet-address"
	DerivedWalletAddress ContextKey = "derived-wallet-address"
	TracerKey            ContextKey = "daoctl-tracer"
	Debug                ContextKey = "debug"
	ConfigVersion                   = "0"
	DefaultNetName                  = "default"
	DefaultKeyFile                  = "default-key"
)

// TODO: Sven just discovered you can use https://pkg.go.dev/github.com/go-logr/logr#NewContext
// TODO: and logr's Marshal could be used to defer the call to spew
//
//	https://pkg.go.dev/github.com/go-logr/logr#Marshaler
type GlobalOptions struct {
	Log logr.Logger
	Ctx context.Context
}

type Publisher struct {
	Protocol      string `yaml:"protocol"`
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	PublishedPort int    `yaml:"published_port"`
	TargetPort    int    `yaml:"target_port"`
}

type WorknetConfig struct {
	KeyFile string      `yaml:"key_file"`
	Ports   []Publisher `yaml:"ports"`
}

type AgentConfig struct {
	Version   string                    `yaml:"version"`
	ActiveNet string                    `yaml:"active"`
	Worknets  map[string]*WorknetConfig `yaml:"worknets"`
}

func LicenseMint(ctx context.Context) gagliardetto.PublicKey {
	cluster := SolanaCluster(ctx)
	if cluster == gagliardettorpc.LocalNet {
		return gagliardetto.MustPublicKeyFromBase58("3CrKoTYzfbeenzQmhXMQzHM929kioTvsr1JtDSX9uET5")
	}
	return gagliardetto.MustPublicKeyFromBase58("Ew5hokTuULRDsgnhKThGv3nrw3RPjiHASQZNcRNTHJ9Z")
}

// TODO: duplicate - see GetClusterByName in endpoints.go
func SolanaCluster(ctx context.Context) gagliardettorpc.Cluster {
	flag := ctx.Value(URL).(string)
	if strings.HasPrefix("mainnet-beta", flag) {
		return gagliardettorpc.MainNetBeta
	} else if strings.HasPrefix("devnet", flag) {
		return gagliardettorpc.DevNet
	} else if strings.HasPrefix("testnet", flag) {
		return gagliardettorpc.TestNet
	} else if strings.HasPrefix("localhost", flag) {
		return gagliardettorpc.LocalNet
	}
	return gagliardettorpc.Cluster{
		Name: flag,
		RPC:  "http://" + flag + ":8899",
		WS:   "ws://" + flag + ":8900",
	}
}

func (g GlobalOptions) Debug() logr.Logger {
	return g.Log.V(1)
}
func (g GlobalOptions) VDebug() logr.Logger {
	return g.Log.V(2)
}

func GetConfigDir(app string) (dir string, err error) {
	// TODO: on windows, this isn't quite right
	// Solana uses C:\Users\svend\.config\solana\id.json
	// os.UserConfigDir gets us C:\Users\svend\AppData\Roaming\solana\id.json
	configDir := "/etc" // GOOD root default for OSX and Linux
	if runtime.GOOS == "windows" {
		// TODO: really want a system path, but we don't support windows yet...
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configDir = filepath.Join(homeDir, ".config")
		}
	}
	if !serviceimpl.Admin() {
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				configDir = filepath.Join(homeDir, ".config")
			}
		} else {
			uConfigDir, err := os.UserConfigDir()
			if err == nil {
				configDir = uConfigDir
			}
		}
	}

	solanaCfgDir := filepath.Join(configDir, app)

	err = os.MkdirAll(solanaCfgDir, os.ModeDir|0700)

	return solanaCfgDir, err
}

func (ac *AgentConfig) Active() (*WorknetConfig, error) {
	worknet, ok := ac.Worknets[ac.ActiveNet]
	if !ok {
		return nil, fmt.Errorf("active network %q not found", ac.ActiveNet)
	}
	return worknet, nil
}

func ValidateAgentConfig(agentConfig *AgentConfig) error {
	activeNetExists := false

	for netName, worknet := range agentConfig.Worknets {
		var portDupeCheck = make(map[int]bool)
		for _, port := range worknet.Ports {
			if _, ok := portDupeCheck[port.PublishedPort]; ok {
				return fmt.Errorf("duplicate port %d", port.PublishedPort)
			}
			portDupeCheck[port.PublishedPort] = true
		}
		if agentConfig.ActiveNet == netName {
			activeNetExists = true
		}
	}

	if !activeNetExists {
		return fmt.Errorf("active network %s does not exist", agentConfig.ActiveNet)
	}

	return nil
}

func WriteAgentConfig(agentConfig *AgentConfig) error {
	if err := ValidateAgentConfig(agentConfig); err != nil {
		return err
	}

	configDir, err := GetConfigDir("WorkNet")
	if err != nil {
		return err
	}

	agentConfigFile := filepath.Join(configDir, "config.yaml")

	if yamlData, err := yaml.Marshal(agentConfig); err != nil {
		return fmt.Errorf("couldn't marshal proxy config: %s", err)
	} else {
		if err := ioutil.WriteFile(agentConfigFile, yamlData, 0644); err != nil {
			return fmt.Errorf("couldn't save proxy config: %s", err)
		}
	}

	return nil
}

func Config() (*AgentConfig, error) {
	configDir, err := GetConfigDir("WorkNet")
	if err != nil {
		return nil, err
	}

	agentConfigFile := filepath.Join(configDir, "config.yaml")

	// check that the file from config dir exists
	// if it doesn't, create it
	if _, err := os.Stat(agentConfigFile); os.IsNotExist(err) {
		config := &AgentConfig{
			Version:   ConfigVersion,
			ActiveNet: DefaultNetName,
			Worknets: map[string]*WorknetConfig{
				DefaultNetName: {
					KeyFile: DefaultKeyFile,
					Ports:   []Publisher{},
				},
			},
		}
		if err := ValidateAgentConfig(config); err != nil {
			return nil, err
		}
		if err := WriteAgentConfig(config); err != nil {
			return nil, err
		}
	}

	configYAML, err := ioutil.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	agentConfig := &AgentConfig{}
	err = yaml.Unmarshal(configYAML, &agentConfig)
	if err != nil {
		return nil, err
	}

	return agentConfig, nil
}
