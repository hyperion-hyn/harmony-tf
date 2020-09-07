package edit

import (
	"fmt"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/microstake"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// NonExistingScenario - executes an edit validator test case using a non-existing validator
func NonExistingScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	fundingMultiple := int64(1)
	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.CreateMap3Node.Map3Node.Amount, fundingMultiple)
	if testCase.ErrorOccurred(err) {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Map3Node")
	account, err := testing.GenerateAndFundAccount(testCase, validatorName, testCase.StakingParameters.CreateMap3Node.Map3Node.Amount, fundingMultiple)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch latest account balance for the account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}

	testCase.StakingParameters.CreateMap3Node.Map3Node.Account = &account
	tx, err := microstake.BasicEditMap3Node(testCase, account.Address, &account, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to edit validator using account %s, address: %s", account.Name, account.Address)
		testCase.HandleError(err, &account, msg)
		return
	}
	testCase.Transactions = append(testCase.Transactions, tx)

	testCase.Result = tx.Success

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testing.Teardown(&account, config.Configuration.Funding.Account.Address)
	testCase.FinishedAt = time.Now().UTC()
}