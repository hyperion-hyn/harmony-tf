package restaking

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/crypto"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	golibMicrostake "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkDelegation "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/delegation"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing"
	"time"
)

// ReuseOrCreateValidator - reuse an existing validator or create a new one
func ReuseOrCreateValidator(testCase *testing.TestCase, validatorName string) (account *sdkAccounts.Account, validator *sdkValidator.Validator, err error) {
	if testCase.StakingParameters.ReuseExistingValidator && config.Configuration.Framework.CurrentValidator != nil {
		return config.Configuration.Framework.CurrentValidator.Account, config.Configuration.Framework.CurrentValidator, nil
	}

	validator = &testCase.StakingParameters.CreateRestaking.Validator
	acc, err := testing.GenerateAndFundAccount(testCase, validatorName, testCase.StakingParameters.CreateRestaking.Validator.Amount, 1)
	if err != nil {
		return nil, nil, err
	}

	account = &acc
	validator.Account = account
	testCase.StakingParameters.CreateRestaking.Validator.Account = account

	testCase.StakingParameters.CreateRestaking.Map3Node.Account = &acc
	map3NodeTx, _, map3NodeExists, err := BasicCreateMap3Node(testCase, account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, account, msg)
		return
	}

	if !map3NodeExists {
		msg := fmt.Sprintf("Create map3Node not exist ")
		testCase.HandleError(err, account, msg)
		return
	}

	golibMicrostake.WaitActive()

	tx, createdBlsKeys, validatorExists, err := BasicCreateValidator(testCase, map3NodeTx.ContractAddress, validator.Account, nil, nil)
	if err != nil {
		return account, nil, err
	}
	testCase.Transactions = append(testCase.Transactions, tx)
	testCase.StakingParameters.CreateRestaking.Validator.BLSKeys = createdBlsKeys
	testCase.StakingParameters.CreateRestaking.Validator.ValidatorAddress = tx.ContractAddress
	testCase.StakingParameters.CreateRestaking.Validator.OperatorAddress = map3NodeTx.ContractAddress

	if config.Configuration.Network.StakingWaitTime > 0 {
		time.Sleep(time.Duration(config.Configuration.Network.StakingWaitTime) * time.Second)
	}

	validator.Exists = validatorExists
	validator.BLSKeys = createdBlsKeys
	validator.ValidatorAddress = tx.ContractAddress
	validator.OperatorAddress = map3NodeTx.ContractAddress

	// The ending balance of the account that created the validator should be less than the funded amount since the create validator tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(validator.Account.Address)
	expectedAccountEndingBalance := account.Balance.Sub(testCase.StakingParameters.CreateRestaking.Validator.Amount)

	if testCase.Expected {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f after creating the validator - expected value: %f (or less)", validator.Account.Name, validator.Account.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)
	} else {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f after creating the validator", validator.Account.Name, validator.Account.Address, accountEndingBalance), testCase.Verbose)
	}

	if testCase.StakingParameters.ReuseExistingValidator && config.Configuration.Framework.CurrentValidator == nil && validatorExists {
		config.Configuration.Framework.CurrentValidator = validator
		validator = config.Configuration.Framework.CurrentValidator
	}

	return account, validator, nil
}

func BasicCreateMap3Node(testCase *testing.TestCase, validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, blsKeys []sdkCrypto.BLSKey) (sdkTxs.Transaction, []sdkCrypto.BLSKey, bool, error) {
	if senderAccount == nil {
		senderAccount = validatorAccount
	}

	if blsKeys == nil || len(blsKeys) == 0 {
		blsKeys = crypto.GenerateBlsKeys(testCase.StakingParameters.CreateRestaking.BLSKeyCount, testCase.StakingParameters.CreateRestaking.BLSSignatureMessage)
	}

	switch testCase.StakingParameters.Mode {
	case "duplicate_bls_key", "duplicateblskey":
		blsKeys = append(blsKeys, blsKeys[0])
	case "amount_larger_than_balance", "amountlargerthanbalance":
		testCase.StakingParameters.CreateRestaking.Map3Node.Amount = testCase.StakingParameters.CreateRestaking.Map3Node.Amount.Mul(ethCommon.NewDec(2))
	}

	if len(blsKeys) > 0 {
		for _, blsKey := range blsKeys {
			logger.StakingLog(fmt.Sprintf("Using BLS key %s to create the map3Node %s", blsKey.PublicKeyHex, validatorAccount.Address), testCase.Verbose)
		}
	}

	logger.TransactionLog(fmt.Sprintf("Sending create map3Node transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := CreateMap3Node(validatorAccount, senderAccount, &testCase.StakingParameters, blsKeys)
	if err != nil {
		return sdkTxs.Transaction{}, nil, false, err
	}

	tx := sdkTxs.ToTransaction(senderAccount.Address, senderAccount.Address, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed create map3Node - address: %s - transaction hash: %s, tx successful: %s", validatorAccount.Address, tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	validatorExists := sdkMap3Node.Exists(rpcClient, tx.ContractAddress)
	addressExistsColoring := logger.ResultColoring(validatorExists, true)
	logger.StakingLog(fmt.Sprintf("Map3Node with address %s exists: %s", tx.ContractAddress, addressExistsColoring), testCase.Verbose)

	return tx, blsKeys, validatorExists, nil
}

func BasicCreateValidator(testCase *testing.TestCase, map3NodeAddress string, validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, blsKeys []sdkCrypto.BLSKey) (sdkTxs.Transaction, []sdkCrypto.BLSKey, bool, error) {
	if senderAccount == nil {
		senderAccount = validatorAccount
	}

	if blsKeys == nil || len(blsKeys) == 0 {
		blsKeys = crypto.GenerateBlsKeys(testCase.StakingParameters.CreateRestaking.BLSKeyCount, testCase.StakingParameters.CreateRestaking.BLSSignatureMessage)
	}

	switch testCase.StakingParameters.Mode {
	case "duplicate_bls_key", "duplicateblskey":
		blsKeys = append(blsKeys, blsKeys[0])
	case "amount_larger_than_balance", "amountlargerthanbalance":
		testCase.StakingParameters.CreateRestaking.Validator.Amount = testCase.StakingParameters.CreateRestaking.Validator.Amount.Mul(ethCommon.NewDec(2))
	}

	if len(blsKeys) > 0 {
		for _, blsKey := range blsKeys {
			logger.StakingLog(fmt.Sprintf("Using BLS key %s to create the validator %s", blsKey.PublicKeyHex, validatorAccount.Address), testCase.Verbose)
		}
	}

	logger.TransactionLog(fmt.Sprintf("Sending create validator transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := CreateValidator(map3NodeAddress, validatorAccount, senderAccount, &testCase.StakingParameters, blsKeys)
	if err != nil {
		return sdkTxs.Transaction{}, nil, false, err
	}

	tx := sdkTxs.ToTransaction(senderAccount.Address, senderAccount.Address, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed create validator - address: %s - transaction hash: %s, tx successful: %s", validatorAccount.Address, tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	validatorExists := sdkValidator.Exists(rpcClient, tx.ContractAddress)
	addressExistsColoring := logger.ResultColoring(validatorExists, true)
	logger.StakingLog(fmt.Sprintf("Validator with address %s exists: %s", tx.ContractAddress, addressExistsColoring), testCase.Verbose)

	return tx, blsKeys, validatorExists, nil
}

// BasicDelegation - helper method to perform delegation
func BasicDelegation(testCase *testing.TestCase, delegatorAccount *sdkAccounts.Account, validatorAddress string, delegatorAddress string, senderAccount *sdkAccounts.Account) (sdkTxs.Transaction, bool, error) {
	logger.StakingLog("Proceeding to perform delegation...", testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending delegation transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := Delegate(delegatorAccount, validatorAddress, delegatorAddress, senderAccount, &testCase.StakingParameters)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}
	tx := sdkTxs.ToTransaction(delegatorAccount.Address, validatorAddress, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed delegation - transaction hash: %s, tx successful: %s", tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	delegations, err := sdkDelegation.ByValidator(rpcClient, validatorAddress)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}

	delegationSucceeded := false
	for _, del := range delegations {
		if del.DelegatorAddress == address.Parse(delegatorAddress) {
			delegationSucceeded = true
			break
		}
	}

	delegationSucceededColoring := logger.ResultColoring(delegationSucceeded, true)
	logger.StakingLog(fmt.Sprintf("Delegation from %s to %s of %f, successful: %s", delegatorAccount.Address, validatorAddress, testCase.StakingParameters.DelegationRestaking.Delegate.Amount, delegationSucceededColoring), testCase.Verbose)

	return tx, delegationSucceeded, nil
}

// BasicUndelegation - helper method to perform undelegation
func BasicUndelegation(testCase *testing.TestCase, delegatorAccount *sdkAccounts.Account, validatorAddress string, delegatorAddress string, senderAccount *sdkAccounts.Account) (sdkTxs.Transaction, bool, error) {
	logger.StakingLog("Proceeding to perform undelegation...", testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending undelegation transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := Undelegate(delegatorAccount, validatorAddress, delegatorAddress, senderAccount, &testCase.StakingParameters)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}
	tx := sdkTxs.ToTransaction(delegatorAccount.Address, validatorAddress, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed undelegation - transaction hash: %s, tx successful: %s", tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	delegations, err := sdkDelegation.ByValidator(rpcClient, validatorAddress)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}

	//undelegationSucceeded := false
	//var undelegationAmount ethCommon.Dec
	//for _, del := range delegations {
	//	if del.DelegatorAddress == address.Parse(delegatorAddress) && del.Undelegation.Amount.Sign() > 0 {
	//		undelegationSucceeded = true
	//		undelegationAmount = ethCommon.NewDecFromBigInt(del.Undelegation.Amount).QuoInt64(params.Ether)
	//		break
	//	}
	//}
	undelegationSucceeded := true
	for _, del := range delegations {
		if del.DelegatorAddress == delegatorAccount.Account.Address {
			undelegationSucceeded = false
			break
		}
	}

	//double check when not release map3Node
	if !undelegationSucceeded {
		for _, del := range delegations {
			if del.DelegatorAddress == address.Parse(delegatorAddress) && del.Undelegation.Amount.Sign() > 0 {
				undelegationSucceeded = true
				break
			}
		}
	}

	undelegationSucceededColoring := logger.ResultColoring(undelegationSucceeded, true)
	logger.StakingLog(fmt.Sprintf("Performed undelegation from validator %s by delegator %ssuccessful: %s", validatorAddress, delegatorAccount.Address, undelegationSucceededColoring), testCase.Verbose)

	return tx, undelegationSucceeded, nil
}

// BasicUndelegation - helper method to perform undelegation
func BasicCollectRestaking(testCase *testing.TestCase, delegatorAccount *sdkAccounts.Account, validatorAddress string, delegatorAddress string, senderAccount *sdkAccounts.Account) (sdkTxs.Transaction, bool, error) {
	logger.StakingLog("Proceeding to perform collect restaking...", testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending collect restaking transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := CollectRestaking(delegatorAccount, validatorAddress, delegatorAddress, senderAccount, &testCase.StakingParameters)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}
	tx := sdkTxs.ToTransaction(delegatorAccount.Address, validatorAddress, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed collect restaking - transaction hash: %s, tx successful: %s", tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	delegation, err := sdkDelegation.ByDelegator(rpcClient, validatorAddress, delegatorAddress)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}

	collectedSucceeded := false
	fmt.Printf("redelegation:%s reward:%s \n", delegatorAddress, delegation.Reward.String())
	if delegation.Reward.Cmp(ethCommon.Big0) == 0 {
		collectedSucceeded = true
	}

	undelegationSucceededColoring := logger.ResultColoring(collectedSucceeded, true)
	logger.StakingLog(fmt.Sprintf("Performed undelegation from validator %s by delegator %ssuccessful: %s", validatorAddress, delegatorAccount.Address, undelegationSucceededColoring), testCase.Verbose)

	return tx, collectedSucceeded, nil
}

func BasicCreateDelegateMap3Node(testCase *testing.TestCase, validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, blsKeys []sdkCrypto.BLSKey) (sdkTxs.Transaction, []sdkCrypto.BLSKey, bool, error) {
	if senderAccount == nil {
		senderAccount = validatorAccount
	}

	if blsKeys == nil || len(blsKeys) == 0 {
		blsKeys = crypto.GenerateBlsKeys(1, "")
	}

	logger.TransactionLog(fmt.Sprintf("Sending create map3Node transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := CreateDelegateMap3Node(validatorAccount, senderAccount, &testCase.StakingParameters, blsKeys)
	if err != nil {
		return sdkTxs.Transaction{}, nil, false, err
	}

	tx := sdkTxs.ToTransaction(senderAccount.Address, senderAccount.Address, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed create map3Node - address: %s - transaction hash: %s, tx successful: %s", validatorAccount.Address, tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	validatorExists := sdkMap3Node.Exists(rpcClient, tx.ContractAddress)
	addressExistsColoring := logger.ResultColoring(validatorExists, true)
	logger.StakingLog(fmt.Sprintf("Map3Node with address %s exists: %s", tx.ContractAddress, addressExistsColoring), testCase.Verbose)

	return tx, blsKeys, validatorExists, nil
}
