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
	timeFormat string = "2006-01-02 15:04:05 UTC"
)

func generateFileName(theTime time.Time, ext string) string {
	return fmt.Sprintf("%s-UTC.%s", utils.FormattedTimeString(theTime), ext)
}

// ExportCSV - exports test suite results as csv
func ExportCSV(results []*testing.TestCase, dismissed []*testing.TestCase, successfulCount int, failedCount int, totalDuration time.Duration) (string, error) {
	records := [][]string{
		{
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
		},
	}

	if len(results) > 0 {
		for _, result := range results {
			records = append(records, csvRow(result))
		}
	}

	if len(dismissed) > 0 {
		records = append(records, emptyRow())
		records = append(records, emptyRow())
		records = append(records, []string{"Dismissed:", "", "", "", "", "", "", "", "", ""})
		for _, skipped := range dismissed {
			records = append(records, dismissedRow(skipped))
		}
	}

	records = append(records, emptyRow())
	records = append(records, emptyRow())
	records = append(records, []string{"", "", "", "", "", "", "", "Summary:", ""})
	records = append(records, []string{"", "", "", "", "", "", "", "Successful:", fmt.Sprintf("%d", successfulCount)})
	records = append(records, []string{"", "", "", "", "", "", "", "Failed:", fmt.Sprintf("%d", failedCount)})
	records = append(records, []string{"", "", "", "", "", "", "", "Dismissed:", fmt.Sprintf("%d", len(dismissed))})
	records = append(records, emptyRow())
	records = append(records, []string{"", "", "", "", "", "", "", "Duration:", totalDuration.String()})

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

	status := ""
	if testCase.Result == testCase.Expected {
		status = "Success"
	} else {
		status = "Failed"
	}

	return []string{
		testCase.Category,
		testCase.Name,
		testCase.Goal,
		fmt.Sprintf("%t", testCase.Executed),
		fmt.Sprintf("%t", testCase.Expected),
		fmt.Sprintf("%t", testCase.Result),
		status,
		startedAtString,
		finishedAtString,
		durationString,
	}
}

func dismissedRow(testCase *testing.TestCase) []string {
	return []string{
		testCase.Category,
		testCase.Name,
		testCase.Goal,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	}
}

func emptyRow() []string {
	return []string{"", "", "", "", "", "", "", "", "", ""}
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
