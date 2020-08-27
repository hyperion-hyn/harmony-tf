package delegation

import (
	"context"
	"encoding/json"
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
func ByDelegator(node string, address string) ([]DelegationInfo, error) {
	return lookupDelegation(node, goSdkRPC.Method.GetDelegationsByDelegator, address)
}

func lookupDelegation(node string, rpcMethod string, address string) ([]DelegationInfo, error) {
	response := DelegationInfoWrapper{}
	delegationInfo := []DelegationInfo{}

	bytes, err := goSdkRPC.RawRequest(rpcMethod, node, []interface{}{address})
	if err != nil {
		return delegationInfo, err
	}

	json.Unmarshal(bytes, &response)
	delegationInfo = response.Result
	if len(delegationInfo) > 0 {
		for _, info := range delegationInfo {
			info.Initialize()
		}
	}

	return delegationInfo, nil
}
