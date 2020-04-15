package edit

import (
	"fmt"

	"github.com/harmony-one/harmony-tf/accounts"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony-tf/logger"
	"github.com/harmony-one/harmony-tf/staking"
	"github.com/harmony-one/harmony-tf/testing"
)

// InvalidAddressScenario - executes an edit validator test case using an invalid address
func InvalidAddressScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	if testCase.ReportError() {
		return
	}

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, validator, err := staking.ReuseOrCreateValidator(testCase, validatorName)
	if err != nil {
		testing.HandleError(testCase, account, fmt.Sprintf("Failed to create validator using account %s", validatorName), err)
		return
	}

	if validator.Exists {
		invalidAccountName := accounts.GenerateTestCaseAccountName(testCase.Name, "InvalidChanger")
		invalidAccount, err := testing.GenerateAndFundAccount(testCase, invalidAccountName, testCase.StakingParameters.Create.Validator.Amount, 1)
		if err != nil {
			testing.HandleError(testCase, validator.Account, fmt.Sprintf("Failed to generate and fund account %s", invalidAccountName), err)
			return
		}

		lastEditTx, lastEditTxErr := staking.BasicEditValidator(testCase, validator.Account, &invalidAccount, nil, nil)
		if lastEditTxErr != nil {
			testing.HandleError(testCase, validator.Account, fmt.Sprintf("Failed to edit validator using account %s, address: %s", invalidAccount.Name, invalidAccount.Address), lastEditTxErr)
			return
		}
		testCase.Transactions = append(testCase.Transactions, lastEditTx)

		testCase.Result = lastEditTx.Success

		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", testCase.Verbose)

		testing.Teardown(&invalidAccount, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
		if !testCase.StakingParameters.ReuseExistingValidator {
			testing.Teardown(validator.Account, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
		}
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)
}
