package create

import (
	"fmt"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/staking"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// AlreadyExistsScenario - executes a create validator test case where the validator has already previously been created
func AlreadyExistsScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(2)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.Create.Validator.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, err := testing.GenerateAndFundAccount(testCase, validatorName, testCase.StakingParameters.Create.Validator.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund account %s", validatorName)
		testCase.HandleError(err, &account, msg)
		return
	}

	testCase.StakingParameters.Create.Validator.Account = &account
	tx, _, validatorExists, err := staking.BasicCreateValidator(testCase, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	// The ending balance of the account that created the validator should be less than the funded amount since the create validator tx should've used the specified amount for self delegation
	accountEndingBalance, err := balances.GetBalance(account.Address)
	if err != nil {
		msg := fmt.Sprintf("Failed to check ending account balance for account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}

	expectedAccountEndingBalance := account.Balance.Sub(testCase.StakingParameters.Create.Validator.Amount)
	logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test - expected value: %f (or less)", account.Name, account.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)

	testCase.Result = tx.Success && accountEndingBalance.LT(expectedAccountEndingBalance) && validatorExists

	// Try to create the exact same validator here
	secondTx, _, secondValidatorExists, err := staking.BasicCreateValidator(testCase, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create second validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, secondTx)

	testCase.Result = testCase.Result && secondTx.Success && secondValidatorExists

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&account, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
