package edit

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

// InvalidAddressScenario - executes an edit validator test case using an invalid address
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

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, validator, err := staking.ReuseOrCreateValidator(testCase, validatorName)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s", validatorName)
		testCase.HandleError(err, account, msg)
		return
	}

	if validator.Exists {
		invalidAccountName := accounts.GenerateTestCaseAccountName(testCase.Name, "InvalidChanger")
		invalidAccount, err := testing.GenerateAndFundAccount(testCase, invalidAccountName, testCase.StakingParameters.Create.Validator.Amount, fundingMultiple)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund account %s", invalidAccountName)
			testCase.HandleError(err, validator.Account, msg)
			return
		}

		lastEditTx, lastEditTxErr := staking.BasicEditValidator(testCase, validator.Account, &invalidAccount, nil, nil)
		if lastEditTxErr != nil {
			msg := fmt.Sprintf("Failed to edit validator using account %s, address: %s", invalidAccount.Name, invalidAccount.Address)
			testCase.HandleError(lastEditTxErr, validator.Account, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, lastEditTx)

		testCase.Result = lastEditTx.Success

		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)

		testing.Teardown(&invalidAccount, config.Configuration.Funding.Account.Address)
		if !testCase.StakingParameters.ReuseExistingValidator {
			testing.Teardown(validator.Account, config.Configuration.Funding.Account.Address)
		}
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}
