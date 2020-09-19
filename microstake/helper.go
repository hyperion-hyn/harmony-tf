package microstake

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/crypto"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	sdkDelegation "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/delegation"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing"
	"time"
)

// ReuseOrCreateMap3Node - reuse an existing map3Node or create a new one
func ReuseOrCreateMap3Node(testCase *testing.TestCase, validatorName string) (account *sdkAccounts.Account, map3Node *sdkMap3Node.Map3Node, err error) {
	if testCase.StakingParameters.ReuseExistingValidator && config.Configuration.Framework.CurrentValidator != nil {
		return config.Configuration.Framework.CurrentValidator.Account, config.Configuration.Framework.CurrentMap3Node, nil
	}

	map3Node = &testCase.StakingParameters.Create.Map3Node
	acc, err := testing.GenerateAndFundAccount(testCase, validatorName, testCase.StakingParameters.Create.Map3Node.Amount, 1)
	if err != nil {
		return nil, nil, err
	}

	account = &acc
	map3Node.Account = account
	testCase.StakingParameters.Create.Map3Node.Account = account

	tx, createdBlsKeys, map3NodeExists, err := BasicCreateMap3Node(testCase, map3Node.Account, nil, nil)
	if err != nil {
		return account, nil, err
	}
	testCase.Transactions = append(testCase.Transactions, tx)
	testCase.StakingParameters.Create.Map3Node.BLSKeys = createdBlsKeys
	testCase.StakingParameters.Create.Map3Node.Map3Address = tx.ContractAddress

	if config.Configuration.Network.StakingWaitTime > 0 {
		time.Sleep(time.Duration(config.Configuration.Network.StakingWaitTime) * time.Second)
	}

	map3Node.Exists = map3NodeExists
	map3Node.BLSKeys = createdBlsKeys
	map3Node.Map3Address = tx.ContractAddress

	// The ending balance of the account that created the map3Node should be less than the funded amount since the create map3Node tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(map3Node.Account.Address)
	expectedAccountEndingBalance := account.Balance.Sub(testCase.StakingParameters.Create.Map3Node.Amount)

	if testCase.Expected {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f after creating the map3Node - expected value: %f (or less)", map3Node.Account.Name, map3Node.Account.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)
	} else {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f after creating the map3Node", map3Node.Account.Name, map3Node.Account.Address, accountEndingBalance), testCase.Verbose)
	}

	if testCase.StakingParameters.ReuseExistingValidator && config.Configuration.Framework.CurrentMap3Node == nil && map3NodeExists {
		config.Configuration.Framework.CurrentMap3Node = map3Node
		map3Node = config.Configuration.Framework.CurrentMap3Node
	}

	return account, map3Node, nil
}

func BasicCreateMap3Node(testCase *testing.TestCase, validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, blsKeys []sdkCrypto.BLSKey) (sdkTxs.Transaction, []sdkCrypto.BLSKey, bool, error) {
	if senderAccount == nil {
		senderAccount = validatorAccount
	}

	if blsKeys == nil || len(blsKeys) == 0 {
		blsKeys = crypto.GenerateBlsKeys(testCase.StakingParameters.Create.BLSKeyCount, testCase.StakingParameters.Create.BLSSignatureMessage)
	}

	switch testCase.StakingParameters.Mode {
	case "duplicate_bls_key", "duplicateblskey":
		blsKeys = append(blsKeys, blsKeys[0])
	case "amount_larger_than_balance", "amountlargerthanbalance":
		testCase.StakingParameters.Create.Map3Node.Amount = testCase.StakingParameters.Create.Map3Node.Amount.Mul(ethCommon.NewDec(2))
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

// BasicEditMap3Node - helper method to edit a map3Node
func BasicEditMap3Node(testCase *testing.TestCase, map3NodeAddress string, senderAccount *sdkAccounts.Account, blsKeyToRemove *sdkCrypto.BLSKey, blsKeyToAdd *sdkCrypto.BLSKey) (sdkTxs.Transaction, error) {
	if senderAccount == nil {
		panic("senderAccount is nil")
	}

	logger.StakingLog(fmt.Sprintf("Proceeding to edit the map3Node %s ...", map3NodeAddress), testCase.Verbose)
	testCase.StakingParameters.EditMap3Node.DetectChanges(testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending edit map3Node transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	editRawTx, err := EditMap3Node(map3NodeAddress, senderAccount, &testCase.StakingParameters, blsKeyToRemove, blsKeyToAdd)
	if err != nil {
		return sdkTxs.Transaction{}, err
	}
	editTx := sdkTxs.ToTransaction(senderAccount.Address, senderAccount.Address, editRawTx, err)
	editTxResultColoring := logger.ResultColoring(editTx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed edit map3Node - transaction hash: %s, tx successful: %s", editTx.TransactionHash, editTxResultColoring), testCase.Verbose)

	return editTx, nil
}

// BasicDelegation - helper method to perform delegation
func BasicDelegation(testCase *testing.TestCase, delegatorAccount *sdkAccounts.Account, map3NodeAddress string, senderAccount *sdkAccounts.Account) (sdkTxs.Transaction, bool, error) {
	logger.StakingLog("Proceeding to perform delegation...", testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending delegation transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := Delegate(delegatorAccount, map3NodeAddress, senderAccount, &testCase.StakingParameters)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}
	tx := sdkTxs.ToTransaction(delegatorAccount.Address, map3NodeAddress, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed delegation - transaction hash: %s, tx successful: %s", tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	delegations, err := sdkDelegation.ByMap3Node(rpcClient, map3NodeAddress)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}

	delegationSucceeded := false
	for _, del := range delegations {
		if del.DelegatorAddress == address.Parse(delegatorAccount.Address) {
			delegationSucceeded = true
			break
		}
	}

	delegationSucceededColoring := logger.ResultColoring(delegationSucceeded, true)
	logger.StakingLog(fmt.Sprintf("DelegationMap3Node from %s to %s of %f, successful: %s", delegatorAccount.Address, map3NodeAddress, testCase.StakingParameters.Delegation.Delegate.Amount, delegationSucceededColoring), testCase.Verbose)

	return tx, delegationSucceeded, nil
}

// BasicUndelegation - helper method to perform undelegation
func BasicUndelegation(testCase *testing.TestCase, delegatorAccount *sdkAccounts.Account, map3NodeAddress string, senderAccount *sdkAccounts.Account) (sdkTxs.Transaction, bool, error) {
	logger.StakingLog("Proceeding to perform undelegation...", testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending undelegation transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := Undelegate(delegatorAccount, map3NodeAddress, senderAccount, &testCase.StakingParameters)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}
	tx := sdkTxs.ToTransaction(delegatorAccount.Address, map3NodeAddress, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed undelegation - transaction hash: %s, tx successful: %s", tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	delegations, err := sdkDelegation.ByMap3Node(rpcClient, map3NodeAddress)
	if err != nil {
		return sdkTxs.Transaction{}, false, err
	}

	undelegationSucceeded := false

	delegateAmount := testCase.StakingParameters.Delegation.Delegate.Amount
	undelegateAmount := testCase.StakingParameters.Delegation.Undelegate.Amount
	remainingDelegateAmount := delegateAmount.Sub(undelegateAmount)

	if undelegateAmount.LT(delegateAmount) {
		// less then delegate amount
		for _, del := range delegations {
			pendingDelegateAmount := ethCommon.NewDecFromBigInt(del.PendingDelegation.Amount).QuoInt64(params.Ether)
			if del.DelegatorAddress == delegatorAccount.Account.Address && pendingDelegateAmount.Equal(remainingDelegateAmount) {
				undelegationSucceeded = true
				break
			}
		}

	} else {
		undelegationSucceeded = true
		for _, del := range delegations {
			if del.DelegatorAddress == delegatorAccount.Account.Address {
				undelegationSucceeded = false
				break
			}
		}

	}

	undelegationSucceededColoring := logger.ResultColoring(undelegationSucceeded, true)
	logger.StakingLog(fmt.Sprintf("Performed undelegation from map3Node %s by delegator %s,expect amount: %f ,successful: %s", map3NodeAddress, delegatorAccount.Address, testCase.StakingParameters.Delegation.Undelegate.Amount, undelegationSucceededColoring), testCase.Verbose)

	return tx, undelegationSucceeded, nil
}

// ManageBLSKeys - manage bls keys for edit map3Node scenarios
func ManageBLSKeys(map3Node *sdkMap3Node.Map3Node, mode string, blsSignatureMessage string, verbose bool) (blsKeyToRemove *sdkCrypto.BLSKey, blsKeyToAdd *sdkCrypto.BLSKey, err error) {
	switch mode {
	case "add_bls_key":
		keyToAdd, err := crypto.GenerateBlsKey(blsSignatureMessage)
		if err != nil {
			fmt.Printf("\n\nStakingParameters.ManageBLSKeys - err: %+v\n\n", err)
			return nil, nil, err
		}
		blsKeyToAdd = &keyToAdd
		logger.StakingLog(fmt.Sprintf("Adding bls key %v to map3Node: %s", blsKeyToAdd.PublicKeyHex, map3Node.Account.Address), verbose)

	case "add_existing_bls_key":
		blsKeyToAdd = &map3Node.BLSKeys[0]
		logger.StakingLog(fmt.Sprintf("Adding already existing bls key %v to map3Node: %s", blsKeyToAdd.PublicKeyHex, map3Node.Account.Address), verbose)

	case "remove_bls_key":
		blsKeyToRemove = &map3Node.BLSKeys[0]
		logger.StakingLog(fmt.Sprintf("Removing bls key %v from map3Node: %s", blsKeyToRemove.PublicKeyHex, map3Node.Account.Address), verbose)

	case "remove_non_existing_bls_key":
		nonExistingKey, err := crypto.GenerateBlsKey(blsSignatureMessage)
		if err != nil {
			return nil, nil, err
		}
		blsKeyToRemove = &nonExistingKey
		logger.StakingLog(fmt.Sprintf("Removing non existing bls key %v from map3Node: %s", blsKeyToRemove.PublicKeyHex, map3Node.Account.Address), verbose)

	case "replace_bls_key":
		keyToAdd, err := crypto.GenerateBlsKey(blsSignatureMessage)
		if err != nil {
			fmt.Printf("\n\nStakingParameters.ManageBLSKeys - err: %+v\n\n", err)
			return nil, nil, err
		}
		blsKeyToAdd = &keyToAdd
		logger.StakingLog(fmt.Sprintf("Adding bls key %v to map3Node: %s", blsKeyToAdd.PublicKeyHex, map3Node.Account.Address), verbose)

		blsKeyToRemove = &map3Node.BLSKeys[0]
		logger.StakingLog(fmt.Sprintf("Removing bls key %v from map3Node: %s", blsKeyToRemove.PublicKeyHex, map3Node.Account.Address), verbose)

	}

	return blsKeyToRemove, blsKeyToAdd, nil
}
