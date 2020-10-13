package restaking

import (
	"errors"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

var (
	errNilAmount             = errors.New("Amount can not be nil")
	errNilMaxTotalDelegation = errors.New("MaxTotalDelegation can not be nil")
)

func CreateValidator(map3NodeAddress string, validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeys []sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if err := validateValidatorValues(params.CreateRestaking.Validator); err != nil {
		return nil, err
	}

	if senderAccount == nil {
		senderAccount = validatorAccount
	}
	senderAccount.Unlock()

	if params.CreateRestaking.Validator.Account == nil {
		params.CreateRestaking.Validator.Account = validatorAccount
	}

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, err
	}

	var currentNonce uint64
	if params.Nonce < 0 {
		currentNonce = sdkNetworkNonce.CurrentNonce(rpcClient, senderAccount.Address)
		if err != nil {
			return nil, err
		}
	} else {
		currentNonce = uint64(params.Nonce)
	}

	txResult, err := sdkValidator.Create(
		senderAccount.Keystore,
		senderAccount.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		map3NodeAddress,
		params.CreateRestaking.Validator.ToStakingDescription(),
		params.CreateRestaking.Validator.ToCommissionRates(),
		params.CreateRestaking.Validator.MaximumTotalDelegation,
		blsKeys,
		params.Gas.Limit,
		params.Gas.Price,
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

func CreateMap3Node(map3NodeAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeys []sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if err := validateMap3NodeValues(params.CreateRestaking.Map3Node); err != nil {
		return nil, err
	}

	if senderAccount == nil {
		senderAccount = map3NodeAccount
	}
	senderAccount.Unlock()

	if params.CreateRestaking.Map3Node.Account == nil {
		params.CreateRestaking.Map3Node.Account = map3NodeAccount
	}

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, err
	}

	var currentNonce uint64
	if params.Nonce < 0 {
		currentNonce = sdkNetworkNonce.CurrentNonce(rpcClient, senderAccount.Address)
		if err != nil {
			return nil, err
		}
	} else {
		currentNonce = uint64(params.Nonce)
	}

	txResult, err := sdkMap3Node.Create(
		senderAccount.Keystore,
		senderAccount.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		map3NodeAccount.Address,
		params.CreateRestaking.Map3Node.ToMicroStakeDescription(),
		params.CreateRestaking.Map3Node.Commission,
		blsKeys,
		params.CreateRestaking.Map3Node.Amount,
		params.Gas.Limit,
		params.Gas.Price,
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

func validateValidatorValues(validator sdkValidator.Validator) error {
	if validator.Amount.IsNil() {
		return errNilAmount
	}

	if validator.MaximumTotalDelegation.IsNil() {
		return errNilMaxTotalDelegation
	}

	return nil
}

func validateMap3NodeValues(map3Node sdkMap3Node.Map3Node) error {
	if map3Node.Amount.IsNil() {
		return errNilAmount
	}

	return nil
}

func CreateDelegateMap3Node(map3NodeAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeys []sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if err := validateMap3NodeValues(params.DelegationRestaking.Delegate.Map3Node); err != nil {
		return nil, err
	}

	if senderAccount == nil {
		senderAccount = map3NodeAccount
	}
	senderAccount.Unlock()

	if params.DelegationRestaking.Delegate.Map3Node.Account == nil {
		params.DelegationRestaking.Delegate.Map3Node.Account = map3NodeAccount
	}

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, err
	}

	var currentNonce uint64
	if params.Nonce < 0 {
		currentNonce = sdkNetworkNonce.CurrentNonce(rpcClient, senderAccount.Address)
		if err != nil {
			return nil, err
		}
	} else {
		currentNonce = uint64(params.Nonce)
	}

	txResult, err := sdkMap3Node.Create(
		senderAccount.Keystore,
		senderAccount.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		map3NodeAccount.Address,
		params.DelegationRestaking.Delegate.Map3Node.ToMicroStakeDescription(),
		params.DelegationRestaking.Delegate.Map3Node.Commission,
		blsKeys,
		params.DelegationRestaking.Delegate.Map3Node.Amount,
		params.Gas.Limit,
		params.Gas.Price,
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
