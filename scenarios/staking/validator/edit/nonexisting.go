package edit

import (
	"fmt"

	"github.com/SebastianJ/harmony-tf/accounts"
	"github.com/SebastianJ/harmony-tf/config"
	"github.com/SebastianJ/harmony-tf/logger"
	"github.com/SebastianJ/harmony-tf/staking"
	"github.com/SebastianJ/harmony-tf/testing"
)

// NonExistingScenario - executes an edit validator test case using a non-existing validator
func NonExistingScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	if testCase.ReportError() {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, err := testing.GenerateAndFundAccount(testCase, validatorName, testCase.StakingParameters.Create.Validator.Amount, 1)
	if err != nil {
		testing.HandleError(testCase, &account, fmt.Sprintf("Failed to fetch latest account balance for the account %s, address: %s", account.Name, account.Address), err)
		return
	}

	testCase.StakingParameters.Create.Validator.Account = &account
	tx, err := staking.BasicEditValidator(testCase, &account, nil, nil, nil)
	if err != nil {
		testing.HandleError(testCase, &account, fmt.Sprintf("Failed to edit validator using account %s, address: %s", account.Name, account.Address), err)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	testCase.Result = tx.Success

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&account, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
}
