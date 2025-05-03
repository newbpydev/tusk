package shared

import (
	"fmt"
	"time"
)

// FormatDueDate returns a formatted due date string based on the current time.
// If the due date is today, it shows the remaining time until the end of the day.
// If it's yesterday, it shows "yesterday" instead of "1 day overdue".
// If it's tomorrow, it shows "tomorrow" instead of just the date.
// If it's overdue by more than 1 day, it shows the number of days overdue.
// Otherwise, it returns the due date in "YYYY-MM-DD" format with appropriate time indications.
// The second return value indicates whether the date is "today", "overdue" or "upcoming" for styling purposes.
func FormatDueDate(due *time.Time, status string) (string, string) {
	if due == nil {
		return "", "upcoming"
	}

	now := time.Now().In(time.Local)
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	taskDueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.Local)

	// If the due date is today, show remaining time until end of day
	if taskDueDate.Equal(todayDate) {
		// Due today - show remaining time until end of day
		endOfDay := todayDate.Add(24 * time.Hour)
		remaining := endOfDay.Sub(now)
		if remaining < 0 {
			remaining = 0
		}
		hours := int(remaining.Hours())
		minutes := int(remaining.Minutes()) % 60
		return fmt.Sprintf("%s (Today: %dh %dm left)", due.Format("2006-01-02"), hours, minutes), "today"
	} else if taskDueDate.Before(todayDate) {
		// Overdue: compute full days overdue
		daysOverdue := int(todayDate.Sub(taskDueDate).Hours() / 24)
		
		// Special case for yesterday (1 day overdue)
		if daysOverdue == 1 {
			return fmt.Sprintf("%s (yesterday)", due.Format("2006-01-02")), "overdue"
		}
		
		return fmt.Sprintf("%s (%d days overdue)", due.Format("2006-01-02"), daysOverdue), "overdue"
	} else {
		// Calculate tomorrow's date for comparison
		tomorrowDate := todayDate.AddDate(0, 0, 1)
		
		// Special case for tomorrow
		if taskDueDate.Equal(tomorrowDate) {
			return fmt.Sprintf("%s (tomorrow)", due.Format("2006-01-02")), "upcoming"
		}
		
		// Calculate days until due
		daysUntil := int(taskDueDate.Sub(todayDate).Hours() / 24)
		
		// For tasks due soon (2-7 days), show the number of days
		if daysUntil <= 7 {
			return fmt.Sprintf("%s (In %d days)", due.Format("2006-01-02"), daysUntil), "upcoming"
		}
		
		return due.Format("2006-01-02"), "upcoming"
	}
}
