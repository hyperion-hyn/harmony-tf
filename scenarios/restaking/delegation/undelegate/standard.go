package undelegate

import (
	"fmt"
	golibMicrostake "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/restaking"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// StandardScenario - executes a standard delegation test case
func StandardScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	requiredFunding := testCase.StakingParameters.CreateRestaking.Validator.Amount.Add(testCase.StakingParameters.DelegationRestaking.Amount)
	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(requiredFunding, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, validator, err := restaking.ReuseOrCreateValidator(testCase, validatorName)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s", validatorName)
		testCase.HandleError(err, account, msg)
		return
	}

	if validator.Exists {
		delegatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Delegator")
		delegatorAccount, err := testing.GenerateAndFundAccount(testCase, delegatorName, testCase.StakingParameters.DelegationRestaking.Amount, 1)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund account %s", delegatorName)
			testCase.HandleError(err, &delegatorAccount, msg)
			return
		}

		testCase.StakingParameters.DelegationRestaking.Delegate.Map3Node.Account = &delegatorAccount

		map3NodeTx, _, map3NodeExists, err := restaking.BasicCreateDelegateMap3Node(testCase, &delegatorAccount, nil, nil)
		if err != nil {
			msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", account.Name, account.Address)
			testCase.HandleError(err, &delegatorAccount, msg)
			return
		}

		if !map3NodeExists {
			msg := fmt.Sprintf("Create map3Node not exist ")
			testCase.HandleError(err, &delegatorAccount, msg)
			return
		}

		golibMicrostake.WaitActive()

		delegationTx, delegationSucceeded, err := restaking.BasicDelegation(testCase, &delegatorAccount, validator.ValidatorAddress, map3NodeTx.ContractAddress, nil)
		if err != nil {
			msg := fmt.Sprintf("Failed to delegate from account %s, address %s to validator %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, validator.Account.Name, validator.Account.Address)
			testCase.HandleError(err, validator.Account, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, delegationTx)

		successfulDelegation := delegationTx.Success && delegationSucceeded

		if successfulDelegation {
			undelegationTx, undelegationSucceeded, err := restaking.BasicUndelegation(testCase, &delegatorAccount, validator.ValidatorAddress, map3NodeTx.ContractAddress, nil)
			if err != nil {
				msg := fmt.Sprintf("Failed to undelegate from account %s, address %s to validator %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, validator.Account.Name, validator.Account.Address)
				testCase.HandleError(err, validator.Account, msg)
				return
			}
			testCase.Transactions = append(testCase.Transactions, undelegationTx)

			testCase.Result = undelegationTx.Success && undelegationSucceeded
		}

		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
		testing.Teardown(&delegatorAccount, config.Configuration.Funding.Account.Address)
	}

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(validator.Account, config.Configuration.Funding.Account.Address)
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}
