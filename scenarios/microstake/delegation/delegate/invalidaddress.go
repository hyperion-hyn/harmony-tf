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
	staking "github.com/hyperion-hyn/hyperion-tf/microstake"
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

	requiredFunding := testCase.StakingParameters.Create.Map3Node.Amount.Add(testCase.StakingParameters.Delegation.Amount)
	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(requiredFunding, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	accounts := map[string]sdkAccounts.Account{}
	accountTypes := []string{
		"map3Node",
		"delegator",
		"sender",
	}

	for _, accountType := range accountTypes {
		accountName := tfAccounts.GenerateTestCaseAccountName(testCase.Name, accountType)
		account, err := testing.GenerateAndFundAccount(testCase, accountName, testCase.StakingParameters.Create.Map3Node.Amount, fundingMultiple)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund %s account %s", accountType, accountName)
			testCase.HandleError(err, &account, msg)
			return
		}
		accounts[accountType] = account
	}

	map3NodeAccount, delegatorAccount, senderAccount := accounts["map3Node"], accounts["delegator"], accounts["sender"]
	testCase.StakingParameters.Create.Map3Node.Account = &map3NodeAccount
	tx, _, map3NodeExists, err := staking.BasicCreateMap3Node(testCase, &map3NodeAccount, &senderAccount, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to create map3Node using account %s, address: %s", map3NodeAccount.Name, map3NodeAccount.Address)
		testCase.HandleError(err, &map3NodeAccount, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	if config.Configuration.Network.StakingWaitTime > 0 {
		time.Sleep(time.Duration(config.Configuration.Network.StakingWaitTime) * time.Second)
	}

	// The ending balance of the account that created the map3Node should be less than the funded amount since the create map3Node tx should've used the specified amount for self delegation
	accountEndingBalance, _ := balances.GetBalance(map3NodeAccount.Address)
	expectedAccountEndingBalance := map3NodeAccount.Balance.Sub(testCase.StakingParameters.Create.Map3Node.Amount)

	if testCase.Expected {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test - expected value: %f (or less)", map3NodeAccount.Name, map3NodeAccount.Address, accountEndingBalance, expectedAccountEndingBalance), testCase.Verbose)
	} else {
		logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has an ending balance of %f  after the test", map3NodeAccount.Name, map3NodeAccount.Address, accountEndingBalance), testCase.Verbose)
	}

	successfulMap3NodeCreation := tx.Success && accountEndingBalance.LT(expectedAccountEndingBalance) && map3NodeExists

	if successfulMap3NodeCreation {
		delegationTx, delegationSucceeded, err := staking.BasicDelegation(testCase, &delegatorAccount, tx.ContractAddress, &senderAccount)
		if err != nil {
			msg := fmt.Sprintf("Failed to delegate from account %s, address %s to map3Node %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, map3NodeAccount.Name, map3NodeAccount.Address)
			testCase.HandleError(err, &map3NodeAccount, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, delegationTx)

		testCase.Result = delegationTx.Success && delegationSucceeded
	}

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)

	testing.Title(testCase, "footer", testCase.Verbose)

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(&map3NodeAccount, config.Configuration.Funding.Account.Address)
	}
	testing.Teardown(&delegatorAccount, config.Configuration.Funding.Account.Address)
	testCase.FinishedAt = time.Now().UTC()
}
