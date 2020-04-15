package testing

import (
	"fmt"

	"github.com/SebastianJ/harmony-tf/config"
	"github.com/gookit/color"
)

// Title - header/footer for test cases
func Title(testCase *TestCase, titleType string, verbose bool) {
	if verbose {
		if titleType == "header" {
			fmt.Println()
		}

		expected := color.Style{color.FgLightWhite, color.BgGreen, color.OpBold}.Render(fmt.Sprintf(" %s ", testCase.ExpectedMessage()))
		executed := ""
		if testCase.Executed {
			executed = config.Configuration.Framework.Styling.TestCaseHeader.Render(" - Result: ")
			resultMsg := ""

			if testCase.Result == testCase.Expected {
				resultMsg = color.Style{color.FgLightWhite, color.BgGreen, color.OpBold}.Render(fmt.Sprintf(" %s ", testCase.ResultMessage()))
			} else {
				resultMsg = color.Style{color.FgLightWhite, color.BgRed, color.OpBold}.Render(fmt.Sprintf(" %s ", testCase.ResultMessage()))
			}
			executed = fmt.Sprintf("%s%s", executed, resultMsg)
		}

		padding := config.Configuration.Framework.Styling.TestCaseHeader.Render(config.Configuration.Framework.Styling.Padding)

		config.Configuration.Framework.Styling.TestCaseHeader.Println(
			fmt.Sprintf("\tTest case: %s - %s: %s - Expected: %s%s%s",
				testCase.Category,
				testCase.Name,
				testCase.Goal,
				expected,
				executed,
				padding,
			),
		)

		if titleType == "footer" {
			fmt.Println()
		}
	}
}
