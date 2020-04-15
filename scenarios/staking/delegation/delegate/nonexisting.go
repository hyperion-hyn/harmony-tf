package delegate

import (
	"fmt"

	"github.com/SebastianJ/harmony-tf/accounts"
	"github.com/SebastianJ/harmony-tf/config"
	"github.com/SebastianJ/harmony-tf/logger"
	"github.com/SebastianJ/harmony-tf/staking"
	"github.com/SebastianJ/harmony-tf/testing"
)

// NonExistingScenario - executes a delegation test case where the validator doesn't exist
func NonExistingScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	if testCase.ReportError() {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	validatorAccount, err := accounts.GenerateAccount(validatorName)
	if err != nil {
		testing.HandleError(testCase, &validatorAccount, fmt.Sprintf("Failed to generate %s account", validatorName), err)
		return
	}

	delegatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Delegator")
	delegatorAccount, err := testing.GenerateAndFundAccount(testCase, delegatorName, testCase.StakingParameters.Delegation.Amount, 1)
	if err != nil {
		testing.HandleError(testCase, &delegatorAccount, fmt.Sprintf("Failed to generate and fund %s account", delegatorName), err)
		return
	}

	delegationTx, delegationSucceeded, err := staking.BasicDelegation(testCase, &delegatorAccount, &validatorAccount, nil)
	if err != nil {
		testing.HandleError(testCase, &validatorAccount, fmt.Sprintf("Failed to delegate from account %s, address %s to validator %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, validatorAccount.Name, validatorAccount.Address), err)
		return
	}
	testCase.Transactions = append(testCase.Transactions, delegationTx)

	testCase.Result = delegationTx.Success && delegationSucceeded

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)

	testing.Title(testCase, "footer", testCase.Verbose)

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(&validatorAccount, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
	}
	testing.Teardown(&delegatorAccount, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
}
