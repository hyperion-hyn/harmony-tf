package edit

import (
	"fmt"

	"github.com/SebastianJ/harmony-tf/accounts"
	"github.com/SebastianJ/harmony-tf/config"
	"github.com/SebastianJ/harmony-tf/logger"
	"github.com/SebastianJ/harmony-tf/staking"
	"github.com/SebastianJ/harmony-tf/testing"
	sdkValidator "github.com/harmony-one/go-lib/staking/validator"
	sdkTxs "github.com/harmony-one/go-lib/transactions"
)

// StandardScenario - executes a standard edit validator test case
func StandardScenario(testCase *testing.TestCase) {
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
		var (
			lastEditTx              sdkTxs.Transaction
			lastValidatorResult     sdkValidator.RPCValidatorResult
			lastSuccessfullyUpdated bool
			lastEditTxErr           error
		)
		node := config.Configuration.Network.API.NodeAddress(testCase.StakingParameters.FromShardID)

		for i := uint32(0); i < testCase.StakingParameters.Edit.Repeat; i++ {
			if i == 0 || (lastEditTxErr == nil && lastEditTx.Success && lastSuccessfullyUpdated) {
				blsKeyToRemove, blsKeyToAdd, blsErr := staking.ManageBLSKeys(validator, testCase.StakingParameters.Edit.Mode, testCase.StakingParameters.Create.BLSSignatureMessage, testCase.Verbose)
				if blsErr != nil {
					testing.HandleError(testCase, validator.Account, fmt.Sprintf("Failed to generate new bls key to use for adding to existing validator %s", validator.Account.Address), blsErr)
					return
				}

				lastEditTx, lastEditTxErr = staking.BasicEditValidator(testCase, validator.Account, nil, blsKeyToRemove, blsKeyToAdd)
				if lastEditTxErr != nil {
					testing.HandleError(testCase, validator.Account, fmt.Sprintf("Failed to edit validator using account %s, address: %s", validator.Account.Name, validator.Account.Address), lastEditTxErr)
					return
				}
				testCase.Transactions = append(testCase.Transactions, lastEditTx)

				lastValidatorResult, lastEditTxErr = sdkValidator.Information(node, validator.Account.Address)
				if lastEditTxErr != nil {
					testing.HandleError(testCase, validator.Account, fmt.Sprintf("Failed to retrieve validator info for validator %s", validator.Account.Address), lastEditTxErr)
					return
				}

				lastSuccessfullyUpdated = testCase.StakingParameters.Edit.EvaluateChanges(lastValidatorResult.Validator, testCase.Verbose)
				editValidatorColoring := logger.ResultColoring(lastSuccessfullyUpdated, true).Render(fmt.Sprintf("%t", lastSuccessfullyUpdated))
				logger.StakingLog(fmt.Sprintf("Validator successfully edited: %s", editValidatorColoring), testCase.Verbose)
			}
		}

		testCase.Result = lastEditTx.Success && lastSuccessfullyUpdated
	}

	if !testCase.StakingParameters.ReuseExistingValidator {
		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", testCase.Verbose)
		testing.Teardown(validator.Account, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)
}
