package create

import (
	"fmt"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/microstake"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// InvalidAddressScenario - executes a create map3Node test case where the map3Node address isn't the same as the account/address sending the create map3Node transaction
func InvalidAddressScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.Create.Map3Node.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	senderName := accounts.GenerateTestCaseAccountName(testCase.Name, "InvalidSender")
	senderAccount, err := testing.GenerateAndFundAccount(testCase, senderName, testCase.StakingParameters.Create.Map3Node.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate and fund account: %s", senderName)
		testCase.HandleError(err, &senderAccount, msg)
		return
	}

	accountName := accounts.GenerateTestCaseAccountName(testCase.Name, "InvalidMap3Node")
	logger.AccountLog(fmt.Sprintf("Generating a new account: %s", accountName), testCase.Verbose)
	map3NodeAccount, err := accounts.GenerateAccount(accountName)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate account %s", accountName)
		testCase.HandleError(err, &map3NodeAccount, msg)
		return
	}
	logger.AccountLog(fmt.Sprintf("Generated account: %s, address: %s", map3NodeAccount.Name, map3NodeAccount.Address), testCase.Verbose)

	testCase.StakingParameters.Create.Map3Node.Account = &map3NodeAccount
	tx, _, map3NodeExists, err := microstake.BasicCreateMap3Node(testCase, &map3NodeAccount, &senderAccount, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create map3Node using account %s, address: %s", senderAccount.Name, senderAccount.Address)
		testCase.HandleError(err, &senderAccount, msg)
		return
	}

	testCase.Transactions = append(testCase.Transactions, tx)

	// The ending balance of the account that created the map3Node should be less than the funded amount since the create map3Node tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(senderAccount.Address)
	expectedAccountEndingBalance := senderAccount.Balance
	logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test - expected value: %f (or less)", senderAccount.Name, senderAccount.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)

	testCase.Result = tx.Success && accountEndingBalance.LT(expectedAccountEndingBalance) && map3NodeExists

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&senderAccount, config.Configuration.Funding.Account.Address)

	testCase.FinishedAt = time.Now().UTC()
}
