package delegation

import (
	"context"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"

	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// ByValidator - get delegations by validator
func ByMap3Node(rpcClient *goSdkRPC.HTTPMessenger, map3NodeAddress string) ([]microstaking.Microdelegation_, error) {

	map3NodeWrapperRPC, err := rpcClient.GetClient().GetMap3NodeInformation(context.Background(), address.Parse(map3NodeAddress), nil)
	if err != nil {
		return nil, err
	}
	return map3NodeWrapperRPC.Microdelegations, nil

}
