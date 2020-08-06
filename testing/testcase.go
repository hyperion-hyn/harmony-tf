package testing

import (
	"fmt"
	"strings"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing/parameters"
)

// TestCase - represents a test case
type TestCase struct {
	Name              string    `yaml:"name"`
	Category          string    `yaml:"category"`
	Goal              string    `yaml:"goal"`
	Priority          int       `yaml:"priority"`
	Execute           bool      `yaml:"execute"`
	Executed          bool      `yaml:"-"`
	Result            bool      `yaml:"result"`
	Expected          bool      `yaml:"expected"`
	StartedAt         time.Time `yaml:"-"`
	FinishedAt        time.Time `yaml:"-"`
	Verbose           bool      `yaml:"verbose"`
	Scenario          string    `yaml:"scenario"`
	Dismissal         string    `yaml:"-"`
	Error             error
	Parameters        parameters.Parameters        `yaml:"parameters"`
	StakingParameters parameters.StakingParameters `yaml:"staking_parameters"`
	Transactions      []sdkTxs.Transaction
	SuccessfulTxCount int64 `yaml:"-"`
	Function          interface{}
}

// Initialize - initializes and converts values for a given test case
func (testCase *TestCase) Initialize() {
	if testCase.Scenario != "" {
		testCase.Scenario = strings.ToLower(testCase.Scenario)
	}

	if testCase.Parameters.RawAmount != "" {
		if err := testCase.Parameters.Initialize(); err != nil {
			testCase.Error = err
			testCase.Result = false
		}
	}

	if testCase.StakingParameters.Create.Validator.RawAmount != "" || testCase.StakingParameters.Edit.Validator.RawAmount != "" || testCase.StakingParameters.Delegation.Delegate.RawAmount != "" || testCase.StakingParameters.Delegation.Undelegate.RawAmount != "" {
		if err := testCase.StakingParameters.Initialize(); err != nil {
			testCase.Error = err
			testCase.Result = false
		}
	}

	if config.Configuration.Network.Timeout > 0 {
		testCase.Parameters.Timeout = config.Configuration.Network.Timeout
		testCase.StakingParameters.Timeout = config.Configuration.Network.Timeout
	}
}

// Duration - how long it took to run the test case
func (testCase *TestCase) Duration() time.Duration {
	if !testCase.StartedAt.IsZero() && !testCase.FinishedAt.IsZero() {
		return testCase.FinishedAt.Sub(testCase.StartedAt)
	}

	return time.Duration(0)
}

// ReportMemoryDismissal - reports memory dismissal for a given test case
func (testCase *TestCase) ReportMemoryDismissal() {
	msg := fmt.Sprintf(
		"Skipping test case %s due to memory restrictions. Your available system memory is %dMB and test case %s requires at least %dMB of memory to properly run",
		testCase.Name,
		config.Configuration.Framework.SystemMemory,
		testCase.Name,
		config.Configuration.Framework.MinimumRequiredMemory,
	)
	testCase.Dismissal = fmt.Sprintf("Test case requires %dMB of memory, total memory available on your system: %dMB", config.Configuration.Framework.MinimumRequiredMemory, config.Configuration.Framework.SystemMemory)
	logger.WarningLog(msg, testCase.Verbose)
	Title(testCase, "footer", testCase.Verbose)
}

// Successful - if the test case result matches the expected result
func (testCase *TestCase) Successful() bool {
	return testCase.Result == testCase.Expected
}

// Status - test case status represented as a string
func (testCase *TestCase) Status() string {
	if testCase.Successful() {
		return "Success"
	}

	return "Failed"
}

// ErrorMessage - return an error message if an error occurred for the test case
func (testCase *TestCase) ErrorMessage() string {
	if testCase.Error != nil {
		return testCase.Error.Error()
	}

	return ""
}

// ExpectedMessage - convert testCase.Expected to a string message
func (testCase *TestCase) ExpectedMessage() string {
	return statusMessage(testCase.Expected)
}

// ResultMessage - convert testCase.Expected to a string message
func (testCase *TestCase) ResultMessage() string {
	return statusMessage(testCase.Result)
}

func statusMessage(status bool) string {
	if status {
		return "SUCCESS"
	}

	return "FAILURE"
}

// SetErrorState - set the error state for a test case based on a given error
func (testCase *TestCase) SetErrorState() {
	testCase.Result = false
	testCase.FinishedAt = time.Now().UTC()
}

// ErrorOccurred - check if an error has occurred for a test case using either a supplied error or if there's already an error registered for the test case
func (testCase *TestCase) ErrorOccurred(err error) bool {
	if err != nil {
		testCase.Error = err
	}

	if testCase.Error != nil {
		testCase.SetErrorState()
		logger.ErrorLog(testCase.Error.Error(), testCase.Verbose)
		logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
		Title(testCase, "footer", testCase.Verbose)
		return true
	}
	return false
}

// HandleError - handle test case errors (log a message, set the result to false and return any eventual funds)
func (testCase *TestCase) HandleError(err error, account *sdkAccounts.Account, message string) {
	if err != nil {
		testCase.Error = err
		testCase.SetErrorState()

		logger.ErrorLog(err.Error(), testCase.Verbose)

		if account != nil {
			logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
			Teardown(account, testCase.StakingParameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
		}

		logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
		Title(testCase, "footer", testCase.Verbose)
	}
}
