package staking

import (
	"errors"
	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

var (
	errNilAmount             = errors.New("Amount can not be nil")
	errNilMinSelfDelegation  = errors.New("MinSelfDelegation can not be nil")
	errNilMaxTotalDelegation = errors.New("MaxTotalDelegation can not be nil")
)

// CreateValidator - creates a validator
func CreateValidator(validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeys []sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if err := validateValidatorValues(params.Create.Validator); err != nil {
		return nil, err
	}

	if senderAccount == nil {
		senderAccount = validatorAccount
	}
	senderAccount.Unlock()

	if params.Create.Validator.Account == nil {
		params.Create.Validator.Account = validatorAccount
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
		validatorAccount.Address,
		params.Create.Validator.ToStakingDescription(),
		params.Create.Validator.ToCommissionRates(),
		params.Create.Validator.MaximumTotalDelegation,
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

// EditValidator - edits a given validator using the provided information
func EditValidator(validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeyToRemove *sdkCrypto.BLSKey, blsKeyToAdd *sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if senderAccount == nil {
		senderAccount = validatorAccount
	}
	senderAccount.Unlock()

	if params.Edit.Validator.Account == nil {
		params.Edit.Validator.Account = validatorAccount
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

	var commissionRate *ethCommon.Dec
	if !params.Edit.Validator.Commission.Rate.IsNil() {
		commissionRate = &params.Edit.Validator.Commission.Rate
	}

	gasLimit := params.Gas.Limit
	gasPrice := params.Gas.Price

	if params.Edit.Gas.RawPrice != "" {
		gasLimit = params.Edit.Gas.Limit
		gasPrice = params.Edit.Gas.Price
	}

	txResult, err := sdkValidator.Edit(
		senderAccount.Keystore,
		senderAccount.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		validatorAccount.Address,
		params.Edit.Validator.ToStakingDescription(),
		commissionRate,
		params.Edit.Validator.MinimumSelfDelegation,
		params.Edit.Validator.MaximumTotalDelegation,
		blsKeyToRemove,
		blsKeyToAdd,
		params.Edit.Validator.EligibilityStatus,
		gasLimit,
		gasPrice,
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

	if validator.MinimumSelfDelegation.IsNil() {
		return errNilMinSelfDelegation
	}

	if validator.MaximumTotalDelegation.IsNil() {
		return errNilMaxTotalDelegation
	}

	return nil
}
