package edit

import (
	"fmt"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/staking"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// StandardScenario - executes a standard edit validator test case
func StandardScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.Create.Validator.Amount, 1)
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
		var (
			lastEditTx              sdkTxs.Transaction
			lastValidatorResult     sdkValidator.RPCValidatorResult
			lastSuccessfullyUpdated bool
			lastEditTxErr           error
		)
		node := config.Configuration.Network.API.NodeAddress()

		for i := uint32(0); i < testCase.StakingParameters.Edit.Repeat; i++ {
			if i == 0 || (lastEditTxErr == nil && lastEditTx.Success && lastSuccessfullyUpdated) {
				blsKeyToRemove, blsKeyToAdd, blsErr := staking.ManageBLSKeys(validator, testCase.StakingParameters.Edit.Mode, testCase.StakingParameters.Create.BLSSignatureMessage, testCase.Verbose)
				if blsErr != nil {
					msg := fmt.Sprintf("Failed to generate new bls key to use for adding to existing validator %s", validator.Account.Address)
					testCase.HandleError(blsErr, validator.Account, msg)
					return
				}

				lastEditTx, lastEditTxErr = staking.BasicEditValidator(testCase, validator.Account, nil, blsKeyToRemove, blsKeyToAdd)
				if lastEditTxErr != nil {
					msg := fmt.Sprintf("Failed to edit validator using account %s, address: %s", validator.Account.Name, validator.Account.Address)
					testCase.HandleError(lastEditTxErr, validator.Account, msg)
					return
				}
				testCase.Transactions = append(testCase.Transactions, lastEditTx)

				lastValidatorResult, lastEditTxErr = sdkValidator.Information(node, validator.Account.Address)
				if lastEditTxErr != nil {
					msg := fmt.Sprintf("Failed to retrieve validator info for validator %s", validator.Account.Address)
					testCase.HandleError(lastEditTxErr, validator.Account, msg)
					return
				}

				lastSuccessfullyUpdated = testCase.StakingParameters.Edit.EvaluateChanges(lastValidatorResult.Validator, testCase.Verbose)
				editValidatorColoring := logger.ResultColoring(lastSuccessfullyUpdated, true)
				logger.StakingLog(fmt.Sprintf("Validator successfully edited: %s", editValidatorColoring), testCase.Verbose)
			}
		}

		testCase.Result = lastEditTx.Success && lastSuccessfullyUpdated
	}

	if !testCase.StakingParameters.ReuseExistingValidator {
		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
		testing.Teardown(validator.Account, config.Configuration.Funding.Account.Address)
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}
