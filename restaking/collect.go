package restaking

import (
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	sdkDelegation "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/delegation"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

// Delegate - performs delegation
func CollectRestaking(delegator *sdkAccounts.Account, validatorAddress string, delegatorAddress string, sender *sdkAccounts.Account, params *testParams.StakingParameters) (map[string]interface{}, error) {

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

	txResult, err := sdkDelegation.Collect(
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

	if err != nil {
		return nil, err
	}

	return txResult, nil
}
