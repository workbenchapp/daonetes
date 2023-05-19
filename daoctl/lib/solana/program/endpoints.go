package program

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/workbenchapp/worknet/daoctl/lib/solana/anchor/generated/worknet"
)

var clusters = map[string]*rpc.Cluster{
	rpc.MainNetBeta.Name: &rpc.MainNetBeta,
	rpc.TestNet.Name:     &rpc.TestNet,
	rpc.DevNet.Name:      &rpc.DevNet,
	rpc.LocalNet.Name:    &rpc.LocalNet,
}
var DefaultCluster = rpc.DevNet.Name

const WORKNET_V1_PROGRAM_ID = "EdUCoDdRnT5HsQ2Ejy3TWMTQP8iUyMQB4WzoNh45pNX9"

var WORKNET_V1_PROGRAM_PUBKEY solana.PublicKey

func init() {
	WORKNET_V1_PROGRAM_PUBKEY, _ = solana.PublicKeyFromBase58(WORKNET_V1_PROGRAM_ID)
	worknet.SetProgramID(WORKNET_V1_PROGRAM_PUBKEY)
}

// TODO: duplicate - see SolanaCluster in options.go
func GetClusterByName(name string) *rpc.Cluster {
	if name == "" {
		name = DefaultCluster
	}
	cluster, ok := clusters[name]
	if !ok {
		cluster = &rpc.Cluster{
			Name: name,
			RPC:  "http://" + name + ":8899",
			WS:   "ws://" + name + ":8900",
		}
		clusters[name] = cluster
	}

	return cluster
}
