package delegate

import (
	"fmt"
	"time"

	tfAccounts "github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/staking"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// InvalidAddressScenario - executes a delegation test case where the delegator address isn't the sender address
func InvalidAddressScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	requiredFunding := testCase.StakingParameters.Create.Validator.Amount.Add(testCase.StakingParameters.Delegation.Amount)
	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(requiredFunding, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	accounts := map[string]sdkAccounts.Account{}
	accountTypes := []string{
		"validator",
		"delegator",
		"sender",
	}

	for _, accountType := range accountTypes {
		accountName := tfAccounts.GenerateTestCaseAccountName(testCase.Name, accountType)
		account, err := testing.GenerateAndFundAccount(testCase, accountName, testCase.StakingParameters.Create.Validator.Amount, fundingMultiple)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund %s account %s", accountType, accountName)
			testCase.HandleError(err, &account, msg)
			return
		}
		accounts[accountType] = account
	}

	validatorAccount, delegatorAccount, senderAccount := accounts["validator"], accounts["delegator"], accounts["sender"]
	testCase.StakingParameters.Create.Validator.Account = &validatorAccount
	tx, _, validatorExists, err := staking.BasicCreateValidator(testCase, &validatorAccount, &senderAccount, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", validatorAccount.Name, validatorAccount.Address)
		testCase.HandleError(err, &validatorAccount, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	if config.Configuration.Network.StakingWaitTime > 0 {
		time.Sleep(time.Duration(config.Configuration.Network.StakingWaitTime) * time.Second)
	}

	// The ending balance of the account that created the validator should be less than the funded amount since the create validator tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(validatorAccount.Address)
	expectedAccountEndingBalance := validatorAccount.Balance.Sub(testCase.StakingParameters.Create.Validator.Amount)

	if testCase.Expected {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test - expected value: %f (or less)", validatorAccount.Name, validatorAccount.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)
	} else {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test", validatorAccount.Name, validatorAccount.Address, accountEndingBalance), testCase.Verbose)
	}

	successfulValidatorCreation := tx.Success && accountEndingBalance.LT(expectedAccountEndingBalance) && validatorExists

	if successfulValidatorCreation {
		delegationTx, delegationSucceeded, err := staking.BasicDelegation(testCase, &delegatorAccount, tx.ContractAddress, delegatorAccount.Address, &senderAccount)
		if err != nil {
			msg := fmt.Sprintf("Failed to delegate from account %s, address %s to validator %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, validatorAccount.Name, validatorAccount.Address)
			testCase.HandleError(err, &validatorAccount, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, delegationTx)

		testCase.Result = delegationTx.Success && delegationSucceeded
	}

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)

	testing.Title(testCase, "footer", testCase.Verbose)

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(&validatorAccount, config.Configuration.Funding.Account.Address)
	}
	testing.Teardown(&delegatorAccount, config.Configuration.Funding.Account.Address)
	testCase.FinishedAt = time.Now().UTC()
}
