package delegation

import (
	"context"
	"github.com/ethereum/go-ethereum/staking/types/restaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"

	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// ByValidator - get delegations by validator
func ByValidator(rpcClient *goSdkRPC.HTTPMessenger, validatorAddress string) ([]restaking.Redelegation_, error) {

	validatorWrapperRPC, err := rpcClient.GetClient().GetValidatorInformation(context.Background(), address.Parse(validatorAddress), nil)
	if err != nil {
		return nil, err
	}
	return validatorWrapperRPC.Redelegations, nil

}

// ByDelegator - get delegations by delegator
func ByDelegator(rpcClient *goSdkRPC.HTTPMessenger, validatorAddress string, delegatorAddress string) (restaking.Redelegation_, error) {
	redelegation, err := rpcClient.GetClient().GetValidatorRedelegation(context.Background(), address.Parse(validatorAddress), address.Parse(delegatorAddress), nil)
	if err != nil {
		return restaking.Redelegation_{}, err
	}
	return *redelegation, nil
}
