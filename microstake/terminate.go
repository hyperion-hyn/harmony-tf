package microstake

import (
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkDelegation "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/delegation"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

func Terminate(operatorAccount *sdkAccounts.Account, map3NodeAddress string, sender *sdkAccounts.Account, params *testParams.StakingParameters) (map[string]interface{}, error) {
	var account *sdkAccounts.Account
	if sender != nil {
		account = sender
	} else {
		account = operatorAccount
	}

	account.Unlock()

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, err
	}

	var currentNonce uint64
	if params.Nonce < 0 {
		currentNonce = sdkNetworkNonce.CurrentNonce(rpcClient, operatorAccount.Address)
		if err != nil {
			return nil, err
		}
	} else {
		currentNonce = uint64(params.Nonce)
	}
	txResult, err := sdkDelegation.Terminate(
		account.Keystore,
		account.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		operatorAccount.Address,
		map3NodeAddress,
		params.Delegation.Delegate.Gas.Limit,
		params.Delegation.Delegate.Gas.Price,
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
