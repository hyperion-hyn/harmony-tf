package undelegate

import (
	"fmt"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/staking"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// NonExistingScenario - executes a delegation test case where the validator doesn't exist
func NonExistingScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.Delegation.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	validatorAccount, err := accounts.GenerateAccount(validatorName)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate account %s", validatorName)
		testCase.HandleError(err, &validatorAccount, msg)
		return
	}

	delegatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Delegator")
	delegatorAccount, err := testing.GenerateAndFundAccount(testCase, delegatorName, testCase.StakingParameters.Delegation.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund account %s", delegatorName)
		testCase.HandleError(err, &delegatorAccount, msg)
		return
	}

	undelegationTx, undelegationSucceeded, err := staking.BasicUndelegation(testCase, &delegatorAccount, &validatorAccount, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to undelegate from account %s, address %s to validator %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, validatorAccount.Name, validatorAccount.Address)
		testCase.HandleError(err, &validatorAccount, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, undelegationTx)

	testCase.Result = undelegationTx.Success && undelegationSucceeded

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(&validatorAccount, config.Configuration.Funding.Account.Address)
	}
	testing.Teardown(&delegatorAccount, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
