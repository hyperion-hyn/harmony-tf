package microstake

import (
	"errors"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	testParams "github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

var (
	errNilAmount             = errors.New("Amount can not be nil")
	errNilMinSelfDelegation  = errors.New("MinSelfDelegation can not be nil")
	errNilMaxTotalDelegation = errors.New("MaxTotalDelegation can not be nil")
)

func CreateMap3Node(map3NodeAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeys []sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if err := validateMap3NodeValues(params.Create.Map3Node); err != nil {
		return nil, err
	}

	if senderAccount == nil {
		senderAccount = map3NodeAccount
	}
	senderAccount.Unlock()

	if params.Create.Map3Node.Account == nil {
		params.Create.Map3Node.Account = map3NodeAccount
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
		params.Create.Map3Node.ToMicroStakeDescription(),
		params.Create.Map3Node.Commission,
		blsKeys,
		params.Create.Map3Node.Amount,
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

// EditMap3Node - edits a given validator using the provided information
func EditMap3Node(validatorAddress string, senderAccount *sdkAccounts.Account, params *testParams.StakingParameters, blsKeyToRemove *sdkCrypto.BLSKey, blsKeyToAdd *sdkCrypto.BLSKey) (map[string]interface{}, error) {
	if senderAccount == nil {
		panic("sender account is nil")
	}
	senderAccount.Unlock()

	//if params.EditMap3Node.Validator.Account == nil {
	//	params.EditMap3Node.Map3Node.Account = validatorAccount
	//}

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

	gasLimit := params.Gas.Limit
	gasPrice := params.Gas.Price

	if params.EditMap3Node.Gas.RawPrice != "" {
		gasLimit = params.EditMap3Node.Gas.Limit
		gasPrice = params.EditMap3Node.Gas.Price
	}

	txResult, err := sdkMap3Node.Edit(
		senderAccount.Keystore,
		senderAccount.Account,
		rpcClient,
		config.Configuration.Network.API.ChainID,
		validatorAddress,
		params.EditMap3Node.Map3Node.ToMicroStakeDescription(),
		blsKeyToRemove,
		blsKeyToAdd,
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

func validateMap3NodeValues(map3Node sdkMap3Node.Map3Node) error {
	if map3Node.Amount.IsNil() {
		return errNilAmount
	}

	return nil
}
