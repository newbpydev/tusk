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
	return time.Parse("2006-01-02", dateStr)
}
