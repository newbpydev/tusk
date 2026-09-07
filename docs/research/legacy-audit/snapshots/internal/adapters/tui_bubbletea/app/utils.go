package app

import (
	"time"
)

// truncateText truncates text to fit within width
func truncateText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	return text[:width-3] + "..."
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b { // Corrected logic: return a if a > b
		return a
	}
	return b
}

// clamp returns a value constrained between min and max
func clamp(value, minVal, maxVal int) int {
	return min(max(value, minVal), maxVal)
}

// formatDate formats a time.Time to a human-friendly string
func formatDate(t time.Time) string {
	if t.IsZero() {
		return "no due date"
	}
	return t.Format("Jan 02, 2006")
}

// parseDate attempts to parse a date string into a time.Time
func parseDate(dateStr string) (time.Time, error) {
	// Try to parse with time component first
	t, err := time.Parse("2006-01-02 15:04", dateStr)
	if err == nil {
		return t, nil
	}
	
	// Fall back to date-only format
	return time.Parse("2006-01-02", dateStr)
}

// isSameDay compares two time.Time values to determine if they represent the same calendar day,
// ignoring time components and timezone differences.
func isSameDay(date1, date2 time.Time) bool {
	// Convert both dates to UTC to avoid timezone issues
	utc1 := date1.UTC()
	utc2 := date2.UTC()

	// Compare year, month, and day only
	return utc1.Year() == utc2.Year() &&
		utc1.Month() == utc2.Month() &&
		utc1.Day() == utc2.Day()
}

// isBeforeDay determines if date1 is strictly before date2 in calendar days,
// ignoring time components and timezone differences.
func isBeforeDay(date1, date2 time.Time) bool {
	// Convert both dates to UTC to avoid timezone issues
	utc1 := date1.UTC()
	utc2 := date2.UTC()

	// Extract date components only
	date1Midnight := time.Date(utc1.Year(), utc1.Month(), utc1.Day(), 0, 0, 0, 0, time.UTC)
	date2Midnight := time.Date(utc2.Year(), utc2.Month(), utc2.Day(), 0, 0, 0, 0, time.UTC)

	// Compare dates by using Unix timestamps at midnight
	return date1Midnight.Unix() < date2Midnight.Unix()
}

// isAfterDay determines if date1 is strictly after date2 in calendar days,
// ignoring time components and timezone differences.
func isAfterDay(date1, date2 time.Time) bool {
	// Convert both dates to UTC to avoid timezone issues
	utc1 := date1.UTC()
	utc2 := date2.UTC()

	// Extract date components only
	date1Midnight := time.Date(utc1.Year(), utc1.Month(), utc1.Day(), 0, 0, 0, 0, time.UTC)
	date2Midnight := time.Date(utc2.Year(), utc2.Month(), utc2.Day(), 0, 0, 0, 0, time.UTC)

	// Compare dates by using Unix timestamps at midnight
	return date1Midnight.Unix() > date2Midnight.Unix()
}
