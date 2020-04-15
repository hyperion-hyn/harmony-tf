package testcases

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SebastianJ/harmony-tf/config"
	"github.com/SebastianJ/harmony-tf/funding"
	"github.com/SebastianJ/harmony-tf/keys"
	stakingDelegationDelegateScenarios "github.com/SebastianJ/harmony-tf/scenarios/staking/delegation/delegate"
	stakingDelegationUndelegateScenarios "github.com/SebastianJ/harmony-tf/scenarios/staking/delegation/undelegate"
	stakingCreateValidatorScenarios "github.com/SebastianJ/harmony-tf/scenarios/staking/validator/create"
	stakingEditValidatorScenarios "github.com/SebastianJ/harmony-tf/scenarios/staking/validator/edit"
	transactionScenarios "github.com/SebastianJ/harmony-tf/scenarios/transactions"
	"github.com/SebastianJ/harmony-tf/testing"
	"github.com/SebastianJ/harmony-tf/utils"
	"github.com/elliotchance/orderedmap"
	"github.com/gookit/color"
)

var (
	// TestCases - contains all test cases that will get executed
	TestCases []*testing.TestCase

	// Results - contains all executed test case results
	Results []*testing.TestCase

	// Dismissed - contains all dismissed test cases
	Dismissed []*testing.TestCase
)

// Execute - executes all registered/identified test cases
func Execute() error {
	header()

	if err := prepare(); err != nil {
		return err
	}

	if len(TestCases) > 0 {
		execute()
		results()
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
	mapping, err := identifyTestCaseFiles(".yml")
	if err != nil {
		return err
	}

	for el := mapping.Front(); el != nil; el = el.Next() {
		for _, testCaseFile := range el.Value.([]string) {
			testCase := &testing.TestCase{}
			err := utils.ParseYaml(testCaseFile, testCase)

			if err == nil {
				testCase.Initialize()
				TestCases = append(TestCases, testCase)
			} else {
				fmt.Println(fmt.Sprintf("Failed to parse test case file: %s - error: %s. Please make sure the test case file is valid YAML!", testCaseFile, err.Error()))
			}
		}
	}

	fmt.Println(fmt.Sprintf("Found a total of %d test case files", len(TestCases)))

	return nil
}

func identifyTestCaseFiles(ext string) (*orderedmap.OrderedMap, error) {
	files := []string{}

	rootPath := filepath.Join(config.Configuration.Framework.BasePath, "testcases")

	if config.Configuration.Framework.Test != "all" {
		rootPath = fmt.Sprintf("%s/%s", rootPath, config.Configuration.Framework.Test)
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	dirs := []string{}
	for _, testCaseFile := range files {
		dir := path.Dir(testCaseFile)
		if !utils.StringSliceContains(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}

	mapping := orderedmap.NewOrderedMap()
	for _, dir := range dirs {
		sectionFiles := []string{}

		for _, testCaseFile := range files {
			if strings.Contains(testCaseFile, dir) {
				sectionFiles = append(sectionFiles, testCaseFile)
			}

			sort.Sort(utils.ByNumericalFilename(sectionFiles))
			mapping.Set(dir, sectionFiles)
		}
	}

	return mapping, nil
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
			default:
				testCase.Executed = false
				fmt.Println(fmt.Sprintf("Please specify a valid test type for your test case %s", testCase.Name))
			}

			if testCase.Executed {
				Results = append(Results, testCase)
			} else {
				Dismissed = append(Dismissed, testCase)
			}
		} else {
			fmt.Println(fmt.Sprintf("\nTest case %s has the execute attribute set to false - make sure to set it to true if you want to execute this test case\n", testCase.Name))
		}
	}
}

func results() {
	config.Configuration.Framework.EndTime = time.Now()
	duration := config.Configuration.Framework.EndTime.Sub(config.Configuration.Framework.StartTime)
	successfulCount := 0
	failedCount := 0

	for _, testCase := range Results {
		if testCase.Result == testCase.Expected {
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
	fmt.Println(strings.Repeat("-", 50))

	if len(Results) > 0 {
		fmt.Println("")
		color.Style{color.OpBold}.Println("Executed test cases:")
		fmt.Println(strings.Repeat("-", 50))
		for _, testCase := range Results {
			if testCase.Result == testCase.Expected {
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

	fmt.Println("")
	color.Style{color.FgBlack, color.BgWhite, color.OpBold}.Println(
		fmt.Sprintf(
			"\tTest suite status - executed a total of %d test case(s)%s",
			len(Results),
			config.Configuration.Framework.Styling.Padding,
		),
	)
}
