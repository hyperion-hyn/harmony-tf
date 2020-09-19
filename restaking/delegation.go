package restaking

import (
	"errors"
	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	sdkDelegation "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/delegation"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

var (
	errNilDelegate   = errors.New("Delegation amount can not be nil or a negative value")
	errNilUndelegate = errors.New("Undelegation amount can not be nil or a negative value")
)

// Delegate - performs delegation
func Delegate(delegator *sdkAccounts.Account, validatorAddress string, delegatorAddress string, sender *sdkAccounts.Account, params *testParams.StakingParameters) (map[string]interface{}, error) {
	return executeDelegationMethod("delegate", delegator, validatorAddress, delegatorAddress, sender, params)
}

// Undelegate - performs undelegation
func Undelegate(delegator *sdkAccounts.Account, validatorAddress string, delegatorAddress string, sender *sdkAccounts.Account, params *testParams.StakingParameters) (map[string]interface{}, error) {
	return executeDelegationMethod("undelegate", delegator, validatorAddress, delegatorAddress, sender, params)
}

func executeDelegationMethod(method string, delegator *sdkAccounts.Account, validatorAddress string, delegatorAddress string, sender *sdkAccounts.Account, params *testParams.StakingParameters) (txResult map[string]interface{}, err error) {
	if err = validateDelegationValues(params); err != nil {
		return nil, err
	}

	var account *sdkAccounts.Account
	if sender != nil {
		account = sender
	} else {
		account = delegator
	}

	account.Unlock()

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, err
	}

	var currentNonce uint64
	if params.Nonce < 0 {
		currentNonce = sdkNetworkNonce.CurrentNonce(rpcClient, delegator.Address)
		if err != nil {
			return nil, err
		}
	} else {
		currentNonce = uint64(params.Nonce)
	}

	if method == "delegate" {
		txResult, err = sdkDelegation.Delegate(
			account.Keystore,
			account.Account,
			rpcClient,
			config.Configuration.Network.API.ChainID,
			delegatorAddress,
			validatorAddress,
			params.DelegationRestaking.Delegate.Gas.Limit,
			params.DelegationRestaking.Delegate.Gas.Price,
			currentNonce,
			config.Configuration.Account.Passphrase,
			config.Configuration.Network.API.NodeAddress(),
			params.Timeout,
		)
	} else if method == "undelegate" {
		txResult, err = sdkDelegation.Undelegate(
			account.Keystore,
			account.Account,
			rpcClient,
			config.Configuration.Network.API.ChainID,
			delegatorAddress,
			validatorAddress,
			params.DelegationRestaking.Undelegate.Gas.Limit,
			params.DelegationRestaking.Undelegate.Gas.Price,
			currentNonce,
			config.Configuration.Account.Passphrase,
			config.Configuration.Network.API.NodeAddress(),
			params.Timeout,
		)
	}

	if err != nil {
		return nil, err
	}

	return txResult, nil
}

func validateDelegationValues(params *testParams.StakingParameters) error {
	if params.DelegationRestaking.Delegate.RawAmount != "" && (params.DelegationRestaking.Delegate.Amount.IsNil() || params.DelegationRestaking.Delegate.Amount.LT(ethCommon.NewDec(0))) {
		return errNilDelegate
	}

	if params.DelegationRestaking.Undelegate.RawAmount != "" && (params.DelegationRestaking.Undelegate.Amount.IsNil() || params.DelegationRestaking.Undelegate.Amount.LT(ethCommon.NewDec(0))) {
		return errNilUndelegate
	}

	return nil
}