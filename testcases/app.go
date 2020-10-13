package testcases

import (
	"fmt"
	microstakeDelegateMap3NodeScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/microstake/delegation/delegate"
	microstakeRenewMap3NodeScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/microstake/delegation/renew"
	microstakeTerminateMap3NodeScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/microstake/delegation/terminate"
	microstakeUndelegateMap3NodeScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/microstake/delegation/undelegate"
	microstakeCreateMap3NodeScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/microstake/map3node/create"
	microstakeEditMap3NodeScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/microstake/map3node/edit"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/export"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/keys"
	restakingDelegationDelegateScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/restaking/delegation/delegate"
	restakingDelegationUndelegateScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/restaking/delegation/undelegate"
	restakingCreateValidatorScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/restaking/validator/create"
	restakingEditValidatorScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/restaking/validator/edit"
	stakingDelegationDelegateScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/staking/delegation/delegate"
	stakingDelegationUndelegateScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/staking/delegation/undelegate"
	stakingCreateValidatorScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/staking/validator/create"
	stakingEditValidatorScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/staking/validator/edit"
	transactionScenarios "github.com/hyperion-hyn/hyperion-tf/scenarios/transactions"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

var (
	// TestCases - contains all test cases that will get executed
	TestCases []*testing.TestCase

	// Results - contains all executed test case results
	Results []*testing.TestCase

	// Dismissed - contains all dismissed test cases
	Dismissed []*testing.TestCase

	// Failed - contains all failed test cases
	Failed []*testing.TestCase
)

// Execute - executes all registered/identified test cases
func Execute() error {
	header()

	if err := prepare(); err != nil {
		return err
	}

	if len(TestCases) > 0 {
		execute()
		successfulCount, failedCount, duration := results()

		switch strings.ToLower(config.Configuration.Export.Format) {
		case "csv":
			csvPath, err := export.ExportCSV(Results, Dismissed, Failed, successfulCount, failedCount, duration)
			if err != nil {
				fmt.Println("Failed to export test case results to CSV")
			} else if csvPath != "" {
				fmt.Printf("Successfully exported test case results to %s\n", csvPath)
			}
		//case "json":
		default:
		}

		footer()
	} else {
		fmt.Println(fmt.Sprintf("Couldn't find any test cases - are you sure you've placed them in the testcases folder?"))
	}

	return nil
}

func header() {
	fmt.Println()
	config.Configuration.Framework.Styling.Header.Println(
		fmt.Sprintf("\tStarting Harmony TF v%s - Network: %s (%s mode) - Nodes: %s%s",
			config.Configuration.Framework.Version,
			strings.Title(config.Configuration.Network.Name),
			strings.ToUpper(config.Configuration.Network.Mode),
			strings.Join(config.Configuration.Network.Nodes[:], ", "),
			strings.Repeat("\t", 15),
		),
	)
}

func load() error {

	if err := loadTestCases(); err != nil {
		return err
	}

	return nil
}

func prepare() (err error) {
	if err = load(); err != nil {
		return err
	}

	accs, err := keys.LoadKeys()
	if err != nil {
		return err
	}

	if err = funding.SetupFundingAccount(accs); err != nil {
		return err
	}

	return nil
}

func execute() {
	for _, testCase := range TestCases {
		if testCase.Execute {
			switch testCase.Scenario {
			case "transactions/standard":
				transactionScenarios.StandardScenario(testCase)
			case "transactions/same_account":
				transactionScenarios.SameAccountScenario(testCase)
			case "transactions/multiple_senders":
				transactionScenarios.MultipleSenderScenario(testCase)
			case "transactions/multiple_receivers_invalid_nonce":
				transactionScenarios.MultipleReceiverInvalidNonceScenario(testCase)
			case "staking/validator/create/standard":
				stakingCreateValidatorScenarios.StandardScenario(testCase)
			case "staking/validator/create/invalid_address":
				stakingCreateValidatorScenarios.InvalidAddressScenario(testCase)
			case "staking/validator/create/already_exists":
				stakingCreateValidatorScenarios.AlreadyExistsScenario(testCase)
			case "staking/validator/create/existing_bls_key":
				stakingCreateValidatorScenarios.ExistingBLSKeyScenario(testCase)
			case "staking/validator/edit/standard":
				stakingEditValidatorScenarios.StandardScenario(testCase)
			case "staking/validator/edit/invalid_address":
				stakingEditValidatorScenarios.InvalidAddressScenario(testCase)
			case "staking/validator/edit/non_existing":
				stakingEditValidatorScenarios.NonExistingScenario(testCase)
			case "staking/delegation/delegate/standard":
				stakingDelegationDelegateScenarios.StandardScenario(testCase)
			case "staking/delegation/delegate/invalid_address":
				stakingDelegationDelegateScenarios.InvalidAddressScenario(testCase)
			case "staking/delegation/delegate/non_existing":
				stakingDelegationDelegateScenarios.NonExistingScenario(testCase)
			case "staking/delegation/undelegate/standard":
				stakingDelegationUndelegateScenarios.StandardScenario(testCase)
			case "staking/delegation/undelegate/invalid_address":
				stakingDelegationUndelegateScenarios.InvalidAddressScenario(testCase)
			case "staking/delegation/undelegate/non_existing":
				stakingDelegationUndelegateScenarios.NonExistingScenario(testCase)
			case "microstake/node/create/standard":
				microstakeCreateMap3NodeScenarios.StandardScenario(testCase)
			case "microstake/node/create/invalid_address":
				microstakeCreateMap3NodeScenarios.InvalidAddressScenario(testCase)
			case "microstake/node/create/existing_bls_key":
				microstakeCreateMap3NodeScenarios.ExistingBLSKeyScenario(testCase)
			case "microstake/node/edit/standard":
				microstakeEditMap3NodeScenarios.StandardScenario(testCase)
			case "microstake/node/edit/invalid_address":
				microstakeEditMap3NodeScenarios.InvalidAddressScenario(testCase)
			case "microstake/node/edit/non_existing":
				microstakeEditMap3NodeScenarios.NonExistingScenario(testCase)
			case "microstake/delegation/delegate/standard":
				microstakeDelegateMap3NodeScenarios.StandardScenario(testCase)
			case "microstake/delegation/delegate/invalid_address":
				microstakeDelegateMap3NodeScenarios.InvalidAddressScenario(testCase)
			case "microstake/delegation/delegate/non_existing":
				microstakeDelegateMap3NodeScenarios.NonExistingScenario(testCase)
			case "microstake/delegation/undelegate/standard":
				microstakeUndelegateMap3NodeScenarios.StandardScenario(testCase)
			case "microstake/delegation/undelegate/invalid_address":
				microstakeUndelegateMap3NodeScenarios.InvalidAddressScenario(testCase)
			case "microstake/delegation/undelegate/non_existing":
				microstakeUndelegateMap3NodeScenarios.NonExistingScenario(testCase)
			case "microstake/delegation/terminate/standard":
				microstakeTerminateMap3NodeScenarios.StandardScenario(testCase)
			case "microstake/delegation/terminate/non_existing":
				microstakeTerminateMap3NodeScenarios.NonExistingScenario(testCase)
			case "microstake/delegation/terminate/invalid_address":
				microstakeTerminateMap3NodeScenarios.InvalidAddressScenario(testCase)
			case "microstake/delegation/renew/standard":
				microstakeRenewMap3NodeScenarios.StandardScenario(testCase)
			case "microstake/delegation/renew/auto":
				microstakeRenewMap3NodeScenarios.AutoRenewScenario(testCase)
			case "microstake/delegation/renew/non_existing":
				microstakeRenewMap3NodeScenarios.NonExistingScenario(testCase)
			case "microstake/delegation/renew/invalid_address":
				microstakeRenewMap3NodeScenarios.InvalidAddressScenario(testCase)
			case "restaking/validator/create/standard":
				restakingCreateValidatorScenarios.StandardScenario(testCase)
			case "restaking/validator/create/invalid_map3_node_address":
				restakingCreateValidatorScenarios.InvalidMap3NodeAddressScenario(testCase)
			case "restaking/validator/create/invalid_address":
				restakingCreateValidatorScenarios.InvalidAddressScenario(testCase)
			case "restaking/validator/create/existing_bls_key":
				restakingCreateValidatorScenarios.ExistingBLSKeyScenario(testCase)
			case "restaking/validator/edit/standard":
				restakingEditValidatorScenarios.StandardScenario(testCase)
			case "restaking/validator/edit/invalid_address":
				restakingEditValidatorScenarios.InvalidAddressScenario(testCase)
			case "restaking/validator/edit/non_existing":
				restakingEditValidatorScenarios.NonExistingScenario(testCase)
			case "restaking/delegation/delegate/standard":
				restakingDelegationDelegateScenarios.StandardScenario(testCase)
			case "restaking/delegation/delegate/invalid_address":
				restakingDelegationDelegateScenarios.InvalidAddressScenario(testCase)
			case "restaking/delegation/delegate/non_existing":
				restakingDelegationDelegateScenarios.NonExistingScenario(testCase)
			case "restaking/delegation/undelegate/standard":
				restakingDelegationUndelegateScenarios.StandardScenario(testCase)
			case "restaking/delegation/undelegate/invalid_address":
				restakingDelegationUndelegateScenarios.InvalidAddressScenario(testCase)
			case "restaking/delegation/undelegate/non_existing":
				restakingDelegationUndelegateScenarios.NonExistingScenario(testCase)
			default:
				testCase.Executed = false
				fmt.Println(fmt.Sprintf("Please specify a valid test type for your test case %s", testCase.Name))
			}

			if testCase.Executed {
				Results = append(Results, testCase)
				if !testCase.Successful() {
					Failed = append(Failed, testCase)
				}
			} else {
				Dismissed = append(Dismissed, testCase)
			}

			if testCase.Error != nil {
				fmt.Println(fmt.Sprintf("Error is not nil, wait %ds for chain finalize", config.Configuration.Network.ErrorWaitTime))
				time.Sleep(time.Duration(config.Configuration.Network.ErrorWaitTime) * time.Second)
			}

		} else {
			fmt.Println(fmt.Sprintf("\nTest case %s has the execute attribute set to false - make sure to set it to true if you want to execute this test case\n", testCase.Name))
		}
	}
}

func results() (successfulCount int, failedCount int, duration time.Duration) {
	config.Configuration.Framework.EndTime = time.Now().UTC()
	duration = config.Configuration.Framework.EndTime.Sub(config.Configuration.Framework.StartTime)
	successfulCount = 0
	failedCount = 0
	dismissedCount := len(Dismissed)

	for _, testCase := range Results {
		if testCase.Successful() {
			successfulCount++
		} else {
			failedCount++
		}
	}

	fmt.Println("")
	color.Style{color.FgBlack, color.BgWhite, color.OpBold}.Println(
		fmt.Sprintf("\tTest suite status - executed a total of %d test case(s) in %v:%s",
			len(Results),
			duration,
			config.Configuration.Framework.Styling.Padding,
		),
	)
	fmt.Println("")

	color.Style{color.OpBold}.Println("Summary:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(fmt.Sprintf("%s %s", config.Configuration.Framework.Styling.Success.Render("Successful:"), color.Style{color.OpBold}.Sprintf("%d", successfulCount)))
	fmt.Println(fmt.Sprintf("%s %s", config.Configuration.Framework.Styling.Error.Render("Failed:"), color.Style{color.OpBold}.Sprintf("%d", failedCount)))
	fmt.Println(fmt.Sprintf("%s %s", config.Configuration.Framework.Styling.Warning.Render("Dismissed:"), color.Style{color.OpBold}.Sprintf("%d", dismissedCount)))
	fmt.Println(strings.Repeat("-", 50))

	if len(Results) > 0 {
		fmt.Println("")
		color.Style{color.OpBold}.Println("Executed test cases:")
		fmt.Println(strings.Repeat("-", 50))
		for _, testCase := range Results {
			if testCase.Successful() {
				fmt.Println(fmt.Sprintf("%s %s", color.Style{color.OpItalic}.Sprintf("Testcase %s:", testCase.Name), config.Configuration.Framework.Styling.Success.Render("success")))
			} else {
				fmt.Println(fmt.Sprintf("%s %s", color.Style{color.OpItalic}.Sprintf("Testcase %s:", testCase.Name), config.Configuration.Framework.Styling.Error.Render("failed")))
			}
		}
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("")
	}

	if len(Dismissed) > 0 {
		fmt.Println("")
		color.Style{color.OpBold}.Println("Test cases that weren't executed/were dismissed:")
		fmt.Println(strings.Repeat("-", 50))
		for _, testCase := range Dismissed {
			fmt.Println(fmt.Sprintf("%s %s", color.Style{color.OpItalic}.Sprintf("Testcase %s - Reason:", testCase.Name), config.Configuration.Framework.Styling.Warning.Render(testCase.Dismissal)))
		}
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("")
		fmt.Printf("Test suite status - a total of %d test case(s) were dismissed", len(Dismissed))
		fmt.Println("")
	}

	return successfulCount, failedCount, duration
}

func footer() {
	fmt.Println("")
	color.Style{color.FgBlack, color.BgWhite, color.OpBold}.Println(
		fmt.Sprintf(
			"\tTest suite status - executed a total of %d test case(s)%s",
			len(Results),
			config.Configuration.Framework.Styling.Padding,
		),
	)
}
