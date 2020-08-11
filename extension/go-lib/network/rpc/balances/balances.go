package balances

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethParams "github.com/ethereum/go-ethereum/params"
	commonTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	goSDK_RPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
	"github.com/pkg/errors"
)

// GetAllShardBalances - gets the balances in all shards for a given address
func GetAllShardBalances(address string, shards map[uint32]string, retry *commonTypes.Retry) (balances map[uint32]ethCommon.Dec, err error) {
	balances = make(map[uint32]ethCommon.Dec)
	params := []interface{}{address, "latest"}

	for shardID, node := range shards {
		balanceRPCReply, err := goSDK_RPC.Request(goSDK_RPC.Method.GetBalance, node, params)
		if err != nil {
			return nil, errors.Wrapf(err, "rpc.Request")
		}

		rpcBalance, _ := balanceRPCReply["result"].(string)
		balance := common.NewDecFromHex(rpcBalance)
		balance = balance.Quo(ethCommon.NewDec(ethParams.Ether))

		balances[shardID] = balance
	}

	return balances, nil
}

// GetShardBalance - gets the balance for a given node, address and shard
func GetShardBalance(address string, shardID uint32, shards map[uint32]string, retry *commonTypes.Retry) (ethCommon.Dec, error) {
	shardBalances, err := GetAllShardBalances(address, shards, retry)
	if err != nil {
		return ethCommon.ZeroDec(), errors.Wrapf(err, "GetShardBalance")
	}

	shardBalance := shardBalances[shardID]

	return shardBalance, nil
}

// GetTotalBalance - gets the total balance across all shards for a given node and address
func GetTotalBalance(address string, shards map[uint32]string, retry *commonTypes.Retry) (ethCommon.Dec, error) {
	shardBalances, err := GetAllShardBalances(address, shards, retry)
	if err != nil {
		return ethCommon.ZeroDec(), errors.Wrapf(err, "GetTotalBalance")
	}

	totalBalance := ethCommon.ZeroDec()

	for _, balance := range shardBalances {
		totalBalance = totalBalance.Add(balance)
	}

	return totalBalance, nil
}
