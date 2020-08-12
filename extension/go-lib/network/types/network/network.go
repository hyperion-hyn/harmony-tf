package network

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"sync"

	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/balances"
	commonRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	commonTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/utils"
	goSDK_common "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	goSDK_RPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// Network - represents a network configuration
type Network struct {
	Name    string
	Mode    string
	Node    string // Node - override any other node settings, if Node is set it will be used as the node address everywhere
	ChainID *goSDK_common.ChainID
	Retry   commonTypes.Retry
	Mutex   sync.Mutex
}

// Initialize - initializes a given network
func (network *Network) Initialize() {
	network.SetChainID()
}

// SetChainID - sets the chain id for a given network
func (network *Network) SetChainID() {
	chainID, err := utils.IdentifyNetworkChainID(network.Name)
	if chainID == nil || err != nil {
		network.ChainID = &goSDK_common.Chain.TestNet
	} else {
		network.ChainID = chainID
	}
}

// NodeAddress - generates a node address given the network's name and mode + the supplied shardID
func (network *Network) NodeAddress() string {
	if network.Node != "" {
		return network.Node
	}
	generated := utils.GenerateNodeAddress(network.Name, network.Mode)
	network.Node = generated
	return network.Node
}

// IdentifyChainID - identifies a chain id given a network name
func (network *Network) IdentifyChainID() (chain *goSDK_common.ChainID, err error) {
	return utils.IdentifyNetworkChainID(network.Name)
}

// GetAllShardBalances - checks the balances in all shards for a given network, mode and address
func (network *Network) GetBalances(address string) (ethCommon.Dec, error) {
	client, err := network.RPCClient()
	if err != nil {
		return ethCommon.ZeroDec(), errors.Wrapf(err, "GetBalances")
	}
	return balances.GetBalance(client, address)
}

// CurrentNonce - gets the current nonce for a given network, mode and address
func (network *Network) CurrentNonce(address string) uint64 {
	rpcClient, err := network.RPCClient()
	if err != nil {
		return 0
	}
	return nonces.CurrentNonce(rpcClient, address)
}

// RPCClient - resolve the RPC/HTTP Messenger to use for remote commands
func (network *Network) RPCClient() (*goSDK_RPC.HTTPMessenger, error) {
	if len(network.Node) > 0 {
		client, err := commonRPC.NewRPCClient(network.Node)
		return client, err
	}
	return nil, nil
}
