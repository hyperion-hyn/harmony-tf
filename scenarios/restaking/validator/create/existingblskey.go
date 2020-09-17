package create

import (
	"fmt"
	"github.com/hyperion-hyn/hyperion-tf/restaking"
	"github.com/hyperion-hyn/hyperion-tf/testing/parameters"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// ExistingBLSKeyScenario - executes a create validator test case using a previously used BLS key
func ExistingBLSKeyScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.CreateRestaking.Validator.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, err := testing.GenerateAndFundAccount(testCase, validatorName, testCase.StakingParameters.CreateRestaking.Validator.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund the account %s", validatorName)
		testCase.HandleError(err, &account, msg)
		return
	}

	testCase.StakingParameters.CreateRestaking.Map3Node.Account = &account
	tx, _, map3NodeExists, err := restaking.BasicCreateMap3Node(testCase, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}

	if !map3NodeExists {
		msg := fmt.Sprintf("Create map3Node not exist ")
		testCase.HandleError(err, &account, msg)
		return
	}

	logger.Log(fmt.Sprintf("sleep %d second for map3Node active", config.Configuration.Network.WaitMap3ActiveTime), true)
	time.Sleep(time.Duration(config.Configuration.Network.WaitMap3ActiveTime) * time.Second)

	testCase.StakingParameters.CreateRestaking.Validator.Account = &account
	tx, blsKeys, validatorExists, err := restaking.BasicCreateValidator(testCase, tx.ContractAddress, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	if tx.Success && validatorExists {
		logger.StakingLog(fmt.Sprintf("Proceeding with trying to create a new validator using the previously used bls key: %s", blsKeys[0].PublicKeyHex), testCase.Verbose)

		duplicateValidatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator_DuplicateBLSKey")
		duplicateAccount, err := testing.GenerateAndFundAccount(testCase, duplicateValidatorName, testCase.StakingParameters.CreateRestaking.Validator.Amount, 1)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund account: %s", duplicateValidatorName)
			testCase.HandleError(err, &duplicateAccount, msg)
			return
		}

		parameters.GenerateMap3NodeUniqueDetails(&testCase.StakingParameters.CreateRestaking.Map3Node.Details)

		if testCase.StakingParameters.Mode == "duplicate_identity" {
			blsKeys = nil
		} else {
			// reinit identity
			err = testCase.StakingParameters.CreateRestaking.Initialize()
			if err != nil {
				msg := fmt.Sprintf("Failed to re initialize testcase")
				testCase.HandleError(err, &duplicateAccount, msg)
				return
			}
		}

		testCase.StakingParameters.CreateRestaking.Map3Node.Account = &duplicateAccount
		tx, _, map3NodeExists, err := restaking.BasicCreateMap3Node(testCase, &duplicateAccount, nil, nil)
		if err != nil {
			msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
			testCase.HandleError(err, &account, msg)
			return
		}

		if !map3NodeExists {
			msg := fmt.Sprintf("Create map3Node not exist ")
			testCase.HandleError(err, &account, msg)
			return
		}

		logger.Log(fmt.Sprintf("sleep %d second for map3Node active", config.Configuration.Network.WaitMap3ActiveTime), true)
		time.Sleep(time.Duration(config.Configuration.Network.WaitMap3ActiveTime) * time.Second)

		testCase.StakingParameters.CreateRestaking.Validator.Account = &duplicateAccount
		duplicateTx, _, duplicateValidatorExists, err := restaking.BasicCreateValidator(testCase, tx.ContractAddress, &duplicateAccount, nil, blsKeys)
		if err != nil {
			msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", duplicateAccount.Name, duplicateAccount.Address)
			testCase.HandleError(err, &account, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, duplicateTx)

		testCase.Result = duplicateTx.Success && duplicateValidatorExists
		testing.Teardown(&duplicateAccount, config.Configuration.Funding.Account.Address)
	}

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&account, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
