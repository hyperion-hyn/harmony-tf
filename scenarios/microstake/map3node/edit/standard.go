package edit

import (
	"fmt"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/microstake"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// StandardScenario - executes a standard edit map3Node test case
func StandardScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	_, _, err := funding.CalculateFundingDetails(testCase.StakingParameters.CreateMap3Node.Map3Node.Amount, 1)
	if testCase.ErrorOccurred(err) {
		return
	}

	accountName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, map3Node, err := microstake.ReuseOrCreateMap3Node(testCase, accountName)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s", accountName)
		testCase.HandleError(err, account, msg)
		return
	}

	if map3Node.Exists {
		var (
			lastEditTx sdkTxs.Transaction
			//lastValidatorResult     sdkMap3Node.RPCValidatorResult
			lastSuccessfullyUpdated bool
			lastEditTxErr           error
		)
		//node := config.Configuration.Network.API.NodeAddress()

		for i := uint32(0); i < testCase.StakingParameters.EditMap3Node.Repeat; i++ {
			if i == 0 || (lastEditTxErr == nil && lastEditTx.Success && lastSuccessfullyUpdated) {
				blsKeyToRemove, blsKeyToAdd, blsErr := microstake.ManageBLSKeys(map3Node, testCase.StakingParameters.EditMap3Node.Mode, testCase.StakingParameters.CreateMap3Node.BLSSignatureMessage, testCase.Verbose)
				if blsErr != nil {
					msg := fmt.Sprintf("Failed to generate new bls key to use for adding to existing map3Node %s", map3Node.Account.Address)
					testCase.HandleError(blsErr, map3Node.Account, msg)
					return
				}
				if blsKeyToAdd != nil {
					testCase.StakingParameters.EditMap3Node.Map3Node.BLSKeys = append(testCase.StakingParameters.EditMap3Node.Map3Node.BLSKeys, *blsKeyToAdd)
				}
				if blsKeyToRemove != nil {
					testCase.StakingParameters.EditMap3Node.Map3Node.BLSKeys = append(testCase.StakingParameters.EditMap3Node.Map3Node.BLSKeys, *blsKeyToRemove)
				}

				lastEditTx, lastEditTxErr = microstake.BasicEditMap3Node(testCase, map3Node.Map3Address, map3Node.Account, blsKeyToRemove, blsKeyToAdd)
				if lastEditTxErr != nil {
					msg := fmt.Sprintf("Failed to edit map3Node using account %s, address: %s", map3Node.Account.Name, map3Node.Account.Address)
					testCase.HandleError(lastEditTxErr, map3Node.Account, msg)
					return
				}
				testCase.Transactions = append(testCase.Transactions, lastEditTx)

				rpcClient, err := config.Configuration.Network.API.RPCClient()
				if err != nil {
					testCase.HandleError(lastEditTxErr, map3Node.Account, "getRpcClient")
					return
				}
				lastMap3Node, lastEditTxErr := sdkMap3Node.Information(rpcClient, address.Parse(map3Node.Map3Address))
				if lastEditTxErr != nil {
					msg := fmt.Sprintf("Failed to retrieve map3Node info for map3Node %s", map3Node.Account.Address)
					testCase.HandleError(lastEditTxErr, map3Node.Account, msg)
					return
				}

				lastSuccessfullyUpdated = testCase.StakingParameters.EditMap3Node.EvaluateChanges(*lastMap3Node, testCase.Verbose)
				editValidatorColoring := logger.ResultColoring(lastSuccessfullyUpdated, true)
				logger.StakingLog(fmt.Sprintf("Validator successfully edited: %s", editValidatorColoring), testCase.Verbose)
			}
		}

		testCase.Result = lastEditTx.Success && lastSuccessfullyUpdated
	}

	if !testCase.StakingParameters.ReuseExistingValidator {
		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
		testing.Teardown(map3Node.Account, config.Configuration.Funding.Account.Address)
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}
