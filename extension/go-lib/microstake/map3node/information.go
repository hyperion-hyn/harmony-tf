package map3node

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

func Information(rpcClient *goSdkRPC.HTTPMessenger, map3NodeAddress common.Address) (*microstaking.Map3NodeWrapperRPC, error) {

	return rpcClient.GetClient().GetMap3NodeInformation(context.Background(), map3NodeAddress, nil)

}
