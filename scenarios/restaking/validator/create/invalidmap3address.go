package create

import (
	"fmt"
	"github.com/hyperion-hyn/hyperion-tf/restaking"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// StandardScenario - executes a standard create validator test case
func InvalidMap3NodeAddressScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	// createMap3Node

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.CreateRestaking.Map3Node.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	accountName := accounts.GenerateTestCaseAccountName(testCase.Name, "Map3Node")
	account, err := testing.GenerateAndFundAccount(testCase, accountName, testCase.StakingParameters.CreateRestaking.Map3Node.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund account %s", accountName)
		testCase.HandleError(err, &account, msg)
		return
	}

	map3Address := "0x462FC87Eb176c04A7369F4D8fa4Ec9a554E46397"

	// createStakingValidator
	testCase.StakingParameters.CreateRestaking.Validator.Account = &account
	tx, _, validatorExists, err := restaking.BasicCreateValidator(testCase, map3Address, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	if config.Configuration.Network.StakingWaitTime > 0 {
		time.Sleep(time.Duration(config.Configuration.Network.StakingWaitTime) * time.Second)
	}

	// The ending balance of the account that created the validator should be less than the funded amount since the create validator tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(account.Address)
	expectedAccountEndingBalance := account.Balance.Sub(testCase.StakingParameters.CreateRestaking.Validator.Amount)

	if testCase.Expected {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test - expected value: %f (or less)", account.Name, account.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)
	} else {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test", account.Name, account.Address, accountEndingBalance), testCase.Verbose)
	}

	testCase.Result = tx.Success && accountEndingBalance.LT(expectedAccountEndingBalance) && validatorExists

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&account, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
