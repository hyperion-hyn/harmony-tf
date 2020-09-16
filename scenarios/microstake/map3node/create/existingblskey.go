package create

import (
	"fmt"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/microstake"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// ExistingBLSKeyScenario - executes a create map3Node test case using a previously used BLS key
func ExistingBLSKeyScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.CreateMap3Node.Map3Node.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	map3NodeName := accounts.GenerateTestCaseAccountName(testCase.Name, "Map3Node")
	account, err := testing.GenerateAndFundAccount(testCase, map3NodeName, testCase.StakingParameters.CreateMap3Node.Map3Node.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund the account %s", map3NodeName)
		testCase.HandleError(err, &account, msg)
		return
	}

	testCase.StakingParameters.CreateMap3Node.Map3Node.Account = &account
	tx, blsKeys, map3NodeExists, err := microstake.BasicCreateMap3Node(testCase, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create map3Node using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	if tx.Success && map3NodeExists {
		logger.StakingLog(fmt.Sprintf("Proceeding with trying to create a new map3Node using the previously used bls key: %s", blsKeys[0].PublicKeyHex), testCase.Verbose)

		duplicateMap3NodeName := accounts.GenerateTestCaseAccountName(testCase.Name, "Map3Node_DuplicateBLSKey")
		duplicateAccount, err := testing.GenerateAndFundAccount(testCase, duplicateMap3NodeName, testCase.StakingParameters.CreateMap3Node.Map3Node.Amount, 1)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund account: %s", duplicateMap3NodeName)
			testCase.HandleError(err, &duplicateAccount, msg)
			return
		}

		if testCase.StakingParameters.Mode == "duplicate_identity" {
			blsKeys = nil
		} else {
			// reinit identity
			err = testCase.StakingParameters.CreateMap3Node.Initialize()
			if err != nil {
				msg := fmt.Sprintf("Failed to re initialize testcase")
				testCase.HandleError(err, &duplicateAccount, msg)
				return
			}
		}

		testCase.StakingParameters.CreateMap3Node.Map3Node.Account = &duplicateAccount
		duplicateTx, _, duplicateMap3NodeExists, err := microstake.BasicCreateMap3Node(testCase, &duplicateAccount, nil, blsKeys)
		if err != nil {
			msg := fmt.Sprintf("Failed to create map3Node using account %s, address: %s", duplicateAccount.Name, duplicateAccount.Address)
			testCase.HandleError(err, &account, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, duplicateTx)

		testCase.Result = duplicateTx.Success && duplicateMap3NodeExists
		testing.Teardown(&duplicateAccount, config.Configuration.Funding.Account.Address)
	}

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&account, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
