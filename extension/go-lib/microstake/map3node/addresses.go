package map3node

import (
	"context"
	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// All - retrieves all validators
func All(rpcClient *goSdkRPC.HTTPMessenger) (addresses []string, err error) {

	addresses, err = rpcClient.GetClient().GetAllMap3NodeAddresses(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	// convert to hynAddress
	//hynAddress := make([]string, len(addresses))
	//for i, addressTemp := range addresses {
	//	hynAddress[i] = common.MustAddressToBech32(common.HexToAddress(addressTemp))
	//}
	return addresses, nil
}

// Exists - checks if a given validator exists
func Exists(rpcClient *goSdkRPC.HTTPMessenger, validatorAddress string) bool {
	allValidators, err := All(rpcClient)
	if err == nil && len(allValidators) > 0 {
		for _, address := range allValidators {
			if address == validatorAddress {
				return true
			}
		}
	}

	return false
}
