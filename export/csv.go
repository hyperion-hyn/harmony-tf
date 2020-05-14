package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony-tf/testing"
	"github.com/harmony-one/harmony-tf/utils"
)

var (
	timeFormat string   = "2006-01-02 15:04:05 UTC"
	headerRow  []string = []string{
		"Category",
		"Name",
		"Goal",
		"Executed",
		"Expected",
		"Result",
		"Status",
		"Started At",
		"Finished At",
		"Duration",
	}
)

func generateFileName(theTime time.Time, ext string) string {
	return fmt.Sprintf("%s-UTC.%s", utils.FormattedTimeString(theTime), ext)
}

// ExportCSV - exports test suite results as csv
func ExportCSV(results []*testing.TestCase, dismissed []*testing.TestCase, failed []*testing.TestCase, successfulCount int, failedCount int, totalDuration time.Duration) (string, error) {
	records := [][]string{headerRow}

	if len(results) > 0 {
		for _, result := range results {
			records = append(records, csvRow(result))
		}
	}

	if len(dismissed) > 0 {
		records = append(records, emptyRow())
		records = append(records, emptyRow())
		records = append(records, titleRow("Dismissed:"))
		for _, skipped := range dismissed {
			records = append(records, dismissedRow(skipped))
		}
	}

	if len(failed) > 0 {
		records = append(records, emptyRow())
		records = append(records, emptyRow())
		records = append(records, titleRow("Failed:"))
		records = append(records, failedHeaders())

		for _, fail := range failed {
			records = append(records, failedRow(fail))
		}
	}

	records = append(records, emptyRow())
	records = append(records, emptyRow())
	records = append(records, summaryRow("Summary:", ""))
	records = append(records, summaryRow("Successful:", fmt.Sprintf("%d", successfulCount)))
	records = append(records, summaryRow("Failed:", fmt.Sprintf("%d", failedCount)))
	records = append(records, summaryRow("Dismissed:", fmt.Sprintf("%d", len(dismissed))))
	records = append(records, emptyRow())
	records = append(records, summaryRow("Duration:", totalDuration.String()))

	filePath, err := writeCSVToFile(records)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func csvRow(testCase *testing.TestCase) []string {
	startedAtString := ""
	if !testCase.StartedAt.IsZero() {
		startedAtString = testCase.StartedAt.Format(timeFormat)
	}

	finishedAtString := ""
	if !testCase.FinishedAt.IsZero() {
		finishedAtString = testCase.FinishedAt.Format(timeFormat)
	}

	duration := testCase.Duration()
	durationString := ""
	if duration.Seconds() > 0.0 {
		durationString = duration.String()
	}

	return []string{
		testCase.Category,
		testCase.Name,
		testCase.Goal,
		fmt.Sprintf("%t", testCase.Executed),
		fmt.Sprintf("%t", testCase.Expected),
		fmt.Sprintf("%t", testCase.Result),
		testCase.Status(),
		startedAtString,
		finishedAtString,
		durationString,
	}
}

func dismissedRow(testCase *testing.TestCase) []string {
	return padRow(
		[]string{
			testCase.Category,
			testCase.Name,
			testCase.Goal,
		},
		"append",
	)
}

func failedHeaders() []string {
	return padRow(
		[]string{
			"Category",
			"Name",
			"Goal",
			"Error Message",
		},
		"append",
	)
}

func failedRow(testCase *testing.TestCase) []string {
	return padRow(
		[]string{
			testCase.Category,
			testCase.Name,
			testCase.Goal,
			testCase.ErrorMessage(),
		},
		"append",
	)
}

func titleRow(title string) []string {
	return padRow([]string{title}, "append")
}

func summaryRow(label string, value string) []string {
	return padRow([]string{label, value}, "prepend")
}

func emptyRow() []string {
	return padRow([]string{}, "append")
}

func padRow(row []string, mode string) []string {
	values := row
	currentLength := len(row)
	padLength := len(headerRow) - currentLength

	if mode == "append" {
		for i := 0; i < padLength; i++ {
			values = append(values, "")
		}
	} else if mode == "prepend" {
		values = []string{}
		for i := 0; i < padLength; i++ {
			values = append(values, "")
		}
		values = append(values, row...)
	}

	return values
}

func writeCSVToFile(records [][]string) (string, error) {
	fileName := generateFileName(config.Configuration.Framework.StartTime, "csv")
	filePath := filepath.Join(config.Configuration.Export.Path, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	csvWriter.WriteAll(records) // calls Flush internally

	if err := csvWriter.Error(); err != nil {
		return "", err
	}

	return filePath, nil
}
