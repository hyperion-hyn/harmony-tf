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

// InvalidAddressScenario - executes a create validator test case where the validator address isn't the same as the account/address sending the create validator transaction
func InvalidAddressScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.Create.Validator.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	senderName := accounts.GenerateTestCaseAccountName(testCase.Name, "InvalidSender")
	senderAccount, err := testing.GenerateAndFundAccount(testCase, senderName, testCase.StakingParameters.Create.Validator.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund account: %s", senderName)
		testCase.HandleError(err, &senderAccount, msg)
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "InvalidValidator")
	logger.AccountLog(fmt.Sprintf("Generating a new account: %s", validatorName), testCase.Verbose)
	validatorAccount, err := accounts.GenerateAccount(validatorName)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate account %s", validatorName)
		testCase.HandleError(err, &validatorAccount, msg)
		return
	}
	logger.AccountLog(fmt.Sprintf("Generated account: %s, address: %s", validatorAccount.Name, validatorAccount.Address), testCase.Verbose)

	testCase.StakingParameters.Create.Validator.Account = &validatorAccount
	tx, _, validatorExists, err := staking.BasicCreateValidator(testCase, &validatorAccount, &senderAccount, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s, address: %s", senderAccount.Name, senderAccount.Address)
		testCase.HandleError(err, &senderAccount, msg)
		return
	}

	testCase.Transactions = append(testCase.Transactions, tx)

	// The ending balance of the account that created the validator should be less than the funded amount since the create validator tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(senderAccount.Address)
	expectedAccountEndingBalance := senderAccount.Balance
	logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test - expected value: %f (or less)", senderAccount.Name, senderAccount.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)

	testCase.Result = tx.Success && accountEndingBalance.LT(expectedAccountEndingBalance) && validatorExists

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&senderAccount, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
