package testing

import (
	"fmt"
	"strings"

	sdkTxs "github.com/harmony-one/go-lib/transactions"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony-tf/logger"
	"github.com/harmony-one/harmony-tf/testing/parameters"
	"github.com/harmony-one/harmony-tf/utils"
)

// TestCase - represents a test case
type TestCase struct {
	Name              string `yaml:"name"`
	Category          string `yaml:"category"`
	Goal              string `yaml:"goal"`
	Priority          int    `yaml:"priority"`
	Execute           bool   `yaml:"execute"`
	Executed          bool   `yaml:"-"`
	Result            bool   `yaml:"result"`
	Expected          bool   `yaml:"expected"`
	Verbose           bool   `yaml:"verbose"`
	Scenario          string `yaml:"scenario"`
	Dismissal         string `yaml:"-"`
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
		testCase.Parameters.Timeout = utils.NetworkTimeoutAdjustment(config.Configuration.Network.Name, testCase.Parameters.Timeout)
	}

	if testCase.StakingParameters.Create.Validator.RawAmount != "" || testCase.StakingParameters.Edit.Validator.RawAmount != "" || testCase.StakingParameters.Delegation.Delegate.RawAmount != "" || testCase.StakingParameters.Delegation.Undelegate.RawAmount != "" {
		if err := testCase.StakingParameters.Initialize(); err != nil {
			testCase.Error = err
			testCase.Result = false
		}
		testCase.StakingParameters.Timeout = utils.NetworkTimeoutAdjustment(config.Configuration.Network.Name, testCase.StakingParameters.Timeout)
	}
}

// SetError - sets the latest error for a given testcase
func (testCase *TestCase) SetError(err error) {
	testCase.Error = err
	testCase.Result = false
	logger.ErrorLog(err.Error(), testCase.Verbose)
	Title(testCase, "footer", testCase.Verbose)
}

// ReportError - report if there's something wrong with the test case before even starting it (e.g. if some params etc are invalid)
func (testCase *TestCase) ReportError() bool {
	if testCase.Error != nil {
		logger.ErrorLog(testCase.Error.Error(), testCase.Verbose)
		testCase.Result = false
		logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)
		Title(testCase, "footer", testCase.Verbose)
		return true
	}
	return false
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
