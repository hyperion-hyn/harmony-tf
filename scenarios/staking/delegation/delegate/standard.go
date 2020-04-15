package delegate

import (
	"fmt"
	"time"

	"github.com/harmony-one/harmony-tf/accounts"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony-tf/logger"
	"github.com/harmony-one/harmony-tf/staking"
	"github.com/harmony-one/harmony-tf/testing"
)

// StandardScenario - executes a standard delegation test case
func StandardScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	if testCase.ReportError() {
		return
	}
	testCase.StartedAt = time.Now().UTC()

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, validator, err := staking.ReuseOrCreateValidator(testCase, validatorName)
	if err != nil {
		testing.HandleError(testCase, account, fmt.Sprintf("Failed to create validator using account %s", validatorName), err)
		return
	}

	if validator.Exists {
		delegatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Delegator")
		delegatorAccount, err := testing.GenerateAndFundAccount(testCase, delegatorName, testCase.StakingParameters.Delegation.Amount, 1)
		if err != nil {
			testing.HandleError(testCase, &delegatorAccount, fmt.Sprintf("Failed to fetch latest account balance for the account %s, address: %s", delegatorAccount.Name, delegatorAccount.Address), err)
			return
		}

		delegationTx, delegationSucceeded, err := staking.BasicDelegation(testCase, &delegatorAccount, validator.Account, nil)
		if err != nil {
			testing.HandleError(testCase, validator.Account, fmt.Sprintf("Failed to delegate from account %s, address %s to validator %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, validator.Account.Name, validator.Account.Address), err)
			return
		}
		testCase.Transactions = append(testCase.Transactions, delegationTx)

		testCase.Result = delegationTx.Success && delegationSucceeded

		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", testCase.Verbose)
		testing.Teardown(&delegatorAccount, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
	}

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(validator.Account, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}
