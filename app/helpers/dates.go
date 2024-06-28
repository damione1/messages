package helpers

import "time"

func FormatDateForHumans(dateStr string) string {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return ""
	}
	// Format the date to a more readable format
	return date.Format("January 2, 2006")
}
