package validator

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/staking/types/restaking"

	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

//Information - get the validator information for a given address
//func Information(node string, validatorAddress common.Address) (RPCValidatorResult, error) {
//	response := RPCValidatorInfoWrapper{}
//	result := RPCValidatorResult{}
//
//	bytes, err := goSdkRPC.RawRequest(goSdkRPC.Method.GetValidatorInformation, node, []interface{}{validatorAddress})
//	if err != nil {
//		return result, err
//	}
//
//	json.Unmarshal(bytes, &response)
//
//	if response.Error.Message != "" {
//		return result, fmt.Errorf("%s (%d)", response.Error.Message, response.Error.Code)
//	}
//
//	result = response.Result
//	result.Initialize()
//
//	return result, nil
//}

func Information(rpcClient *goSdkRPC.HTTPMessenger, validatorAddress common.Address) (*restaking.PlainValidatorWrapper, error) {

	return rpcClient.GetClient().GetValidatorInformation(context.Background(), validatorAddress, nil)

}
