package microstake

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkDelegation "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/delegation"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

func Renew(delegatorAccount *sdkAccounts.Account, map3NodeAddress string, sender *sdkAccounts.Account, isOperator bool, params *testParams.StakingParameters) (map[string]interface{}, error) {
	var account *sdkAccounts.Account
	if sender != nil {
		account = sender
	} else {
		account = delegatorAccount
	}

	account.Unlock()

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, err
	}

	var currentNonce uint64
	if params.Nonce < 0 {
		currentNonce = sdkNetworkNonce.CurrentNonce(rpcClient, delegatorAccount.Address)
		if err != nil {
			return nil, err
		}
	} else {
		currentNonce = uint64(params.Nonce)
	}

	var isRenew bool

	var newCommissionRate *ethCommon.Dec

	if isOperator {
		isRenew = params.Delegation.Renew.OperatorRenew
		newCommissionRate = params.Delegation.Renew.OperatorCommission
	} else {
		isRenew = params.Delegation.Renew.ParticipantRenew
		newCommissionRate = params.Delegation.Renew.ParticipantCommission
	}

	txResult, err := sdkDelegation.Renew(
		account.Keystore,
		account.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		delegatorAccount.Address,
		map3NodeAddress,
		isRenew,
		newCommissionRate,
		params.Delegation.Renew.Gas.Limit,
		params.Delegation.Renew.Gas.Price,
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
