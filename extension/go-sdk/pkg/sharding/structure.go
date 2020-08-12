package sharding

import (
	"encoding/json"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

var (
	nanoAsDec = ethCommon.NewDec(params.GWei)
	oneAsDec  = ethCommon.NewDec(params.Ether)
)

// RPCRoutes reflects the RPC endpoints of the target network across shards
type RPCRoutes struct {
	HTTP    string `json:"http"`
	ShardID int    `json:"shardID"`
	WS      string `json:"ws"`
}

// Structure produces a slice of RPCRoutes for the network across shards
func Structure(node string) ([]RPCRoutes, error) {
	type r struct {
		Result []RPCRoutes `json:"result"`
	}
	p, e := rpc.RawRequest(rpc.Method.GetShardingStructure, node, []interface{}{})
	if e != nil {
		return nil, e
	}
	result := r{}
	if err := json.Unmarshal(p, &result); err != nil {
		return nil, err
	}
	return result.Result, nil
}
