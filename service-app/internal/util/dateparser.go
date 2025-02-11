package util

import (
	"errors"
	"strings"
	"time"
)

// Define the Go layouts based on the PHP date formats provided.
var dateFormats = []string{
	"02/01/2006",          // d/m/Y
	"02/01/2006 15:04:05", // d/m/Y H:i:s
	"02-01-2006",          // d-m-Y
	"02-01-2006 15:04:05", // d-m-Y H:i:s
	"2006-01-02",          // Y-m-d
	"2006-01-02 15:04:05", // Y-m-d H:i:s
	"2006-01-02T15:04:05", // Y-m-d\TH:i:s
	"02012006",            // dmY (Banktec format)
	"02/01/2006",          // d/m/Y (spreadsheet format)
	"2 January 2006",      // j F Y (Outgoing Correspondence format)
}

// Attempts to parse a date string against multiple formats.
// It returns a parsed time.Time if successful, or an error if no format matches.
func ParseDate(input string, format string) (time.Time, error) {
	if format != "" {
		return time.Parse(format, input)
	}

	input = strings.TrimSpace(input)

	for _, format := range dateFormats {
		if date, err := time.Parse(format, input); err == nil {
			return date, nil
		}
	}

	// If no format matched, return an error
	return time.Time{}, errors.New("input was not in an expected date format")
}
