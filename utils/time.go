package utils

import (
	"strings"
	"time"
)

var (
	timeFormat       string = "2006-01-02 15:04:05 UTC"
	timeFormatString string = "2006-01-02-15-04-05"
)

// FormattedTimeString - generate a time string suitable e.g. for file names etc
func FormattedTimeString(theTime time.Time) string {
	timestamp := theTime.Format(timeFormatString)
	timestamp = strings.ReplaceAll(timestamp, "+", "")
	timestamp = strings.ReplaceAll(timestamp, " ", "-")
	timestamp = strings.ReplaceAll(timestamp, ":", "-")

	return timestamp
}
