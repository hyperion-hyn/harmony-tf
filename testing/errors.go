package testing

import (
	"fmt"

	sdkAccounts "github.com/harmony-one/go-lib/accounts"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony-tf/logger"
)

// HandleError - handle test case errors (log a message, set the result to false and return any eventual funds)
func HandleError(testCase *TestCase, account *sdkAccounts.Account, message string, err error) {
	if err != nil {
		logger.ErrorLog(fmt.Sprintf("%s - error: %s", message, err.Error()), testCase.Verbose)
		testCase.Error = err
		testCase.Result = false
		logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", testCase.Verbose)
		logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
		Title(testCase, "footer", testCase.Verbose)

		if account != nil {
			Teardown(account, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
		}
	}
}
