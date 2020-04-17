package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/harmony-one/harmony-tf/config"
)

var (
	timeFormat = "2006-01-02 15:04:05"
)

// Log - logs default testing messages
func Log(message string, verbose bool) {
	OutputLog(message, "default", verbose)
}

// AccountLog - logs account related testing messages
func AccountLog(message string, verbose bool) {
	OutputLog(message, "account", verbose)
}

// FundingLog - logs funding related testing messages
func FundingLog(message string, verbose bool) {
	OutputLog(message, "funding", verbose)
}

// BalanceLog - logs balance related testing messages
func BalanceLog(message string, verbose bool) {
	OutputLog(message, "balance", verbose)
}

// TransactionLog - logs transaction related testing messages
func TransactionLog(message string, verbose bool) {
	OutputLog(message, "transaction", verbose)
}

// StakingLog - logs staking related testing messages
func StakingLog(message string, verbose bool) {
	OutputLog(message, "staking", verbose)
}

// TeardownLog - logs teardown related testing messages
func TeardownLog(message string, verbose bool) {
	OutputLog(message, "teardown", verbose)
}

// WarningLog - logs error related testing messages
func WarningLog(message string, verbose bool) {
	OutputLog(message, "warning", verbose)
}

// ErrorLog - logs error related testing messages
func ErrorLog(message string, verbose bool) {
	OutputLog(message, "error", verbose)
}

// ResultLog - logs result related testing messages - will switch between green (successful) and red (failed) depending on the passed boolean
func ResultLog(result bool, expected bool, verbose bool) {
	if verbose {
		var formattedCategory string
		message := fmt.Sprintf("Test successful: %t, Expected: %t", result, expected)
		formattedMessage := ResultColoring(result, expected).Render(message)

		if result == expected {
			formattedCategory = color.Style{color.FgGreen, color.OpBold}.Render("RESULT")
		} else {
			formattedCategory = color.Style{color.FgRed, color.OpBold}.Render("RESULT")
		}

		fmt.Println(fmt.Sprintf("\n[%s] %s - %s", time.Now().Format(timeFormat), formattedCategory, formattedMessage))
	}
}

// ResultColoring - generate a green or red color setup depending on if the result matches the expected result
func ResultColoring(result bool, expected bool) *color.Style {
	if result == expected {
		return config.Configuration.Framework.Styling.Success
	}

	return config.Configuration.Framework.Styling.Error
}

// OutputLog - time stamped logging messages for test cases
func OutputLog(message string, category string, verbose bool) {
	if verbose {
		var c *color.Style

		switch category {
		case "default":
			c = config.Configuration.Framework.Styling.Default
		case "account":
			c = config.Configuration.Framework.Styling.Account
		case "funding":
			c = config.Configuration.Framework.Styling.Funding
		case "balance":
			c = config.Configuration.Framework.Styling.Balance
		case "transaction":
			c = config.Configuration.Framework.Styling.Transaction
		case "staking":
			c = config.Configuration.Framework.Styling.Staking
		case "teardown":
			c = config.Configuration.Framework.Styling.Teardown
		case "warning":
			c = config.Configuration.Framework.Styling.Warning
		case "error":
			c = config.Configuration.Framework.Styling.Error
		default:
			c = config.Configuration.Framework.Styling.Default
		}

		formattedCategory := c.Render(strings.ToUpper(category))
		title := fmt.Sprintf("[%s] %s - ", time.Now().Format(timeFormat), formattedCategory)
		fmt.Println(fmt.Sprintf("%s%s", title, message))
	}
}
