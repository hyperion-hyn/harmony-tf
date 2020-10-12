package undelegate

import (
	"fmt"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/utils"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	staking "github.com/hyperion-hyn/hyperion-tf/microstake"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

// StandardScenario - executes a standard delegationMap3Node test case
func AutoRenewScenario(testCase *testing.TestCase) {
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

	validatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Validator")
	account, map3Node, err := staking.ReuseOrCreateMap3Node(testCase, validatorName)
	if err != nil {
		msg := fmt.Sprintf("Failed to create validator using account %s", validatorName)
		testCase.HandleError(err, account, msg)
		return
	}

	if map3Node.Exists {
		delegatorName := accounts.GenerateTestCaseAccountName(testCase.Name, "Delegator")
		delegatorAccount, err := testing.GenerateAndFundAccount(testCase, delegatorName, testCase.StakingParameters.Delegation.Amount, 1)
		if err != nil {
			msg := fmt.Sprintf("Failed to generate and fund account %s", delegatorName)
			testCase.HandleError(err, &delegatorAccount, msg)
			return
		}

		delegationTx, delegationSucceeded, err := staking.BasicDelegation(testCase, &delegatorAccount, map3Node.Map3Address, nil)
		if err != nil {
			msg := fmt.Sprintf("Failed to delegate from account %s, address %s to map3Node %s, address: %s", delegatorAccount.Name, delegatorAccount.Address, map3Node.Account.Name, map3Node.Account.Address)
			testCase.HandleError(err, map3Node.Account, msg)
			return
		}
		testCase.Transactions = append(testCase.Transactions, delegationTx)
		successfulDelegation := delegationTx.Success && delegationSucceeded
		if successfulDelegation {
			rpcClient, err := config.Configuration.Network.API.RPCClient()
			err = utils.WaitForEpoch(rpcClient, 1)
			if err != nil {
				msg := fmt.Sprintf("Wait for skip epoch error")
				testCase.HandleError(err, &delegatorAccount, msg)
				return
			}
			beforeNodeInfo, err := sdkMap3Node.Information(rpcClient, address.Parse(map3Node.Map3Address))
			if err != nil {
				msg := fmt.Sprintf("Failed to get map3Node %s status", map3Node.Map3Address)
				testCase.HandleError(err, map3Node.Account, msg)
				return
			}
			err = utils.WaitForEpoch(rpcClient, microstaking.Map3NodeLockDurationInEpoch.TruncateInt64())
			if err != nil {
				msg := fmt.Sprintf("Wait for skip epoch error")
				testCase.HandleError(err, &delegatorAccount, msg)
				return
			}
			afterNodeInfo, err := sdkMap3Node.Information(rpcClient, address.Parse(map3Node.Map3Address))
			if err != nil {
				msg := fmt.Sprintf("Failed to get map3Node %s status", map3Node.Map3Address)
				testCase.HandleError(err, map3Node.Account, msg)
				return
			}

			releaseDiffEpoch := afterNodeInfo.Map3Node.ReleaseEpoch.Sub(beforeNodeInfo.Map3Node.ReleaseEpoch)
			fmt.Printf("releaseDiffEpoch:%s", releaseDiffEpoch.String())
			renewReleaseEpochSuccessed := releaseDiffEpoch.Equal(microstaking.Map3NodeLockDurationInEpoch)
			renewSucceeded := afterNodeInfo.Map3Node.Status == byte(microstaking.Active) && renewReleaseEpochSuccessed
			delegationSucceededColoring := logger.ResultColoring(renewSucceeded, true)
			logger.StakingLog(fmt.Sprintf("renew  map3Node  %s  , successful: %s", map3Node.Map3Address, delegationSucceededColoring), testCase.Verbose)

			testCase.Transactions = append(testCase.Transactions, delegationTx)
			testCase.Result = delegationTx.Success && renewSucceeded
		}

		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
		testing.Teardown(&delegatorAccount, config.Configuration.Funding.Account.Address)
	}

	if !testCase.StakingParameters.ReuseExistingValidator {
		testing.Teardown(map3Node.Account, config.Configuration.Funding.Account.Address)
	}

	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}
