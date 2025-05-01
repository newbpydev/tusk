package shared

import (
	"fmt"
	"math"
	"time"
)

// FormatDueDate returns a formatted due date string based on the current time.
// If the due date is today, it shows the remaining time until the end of the day.
// If it's overdue, it shows the number of days overdue.
// Otherwise, it returns the due date in "YYYY-MM-DD" format.
// The second return value indicates whether the date is "today", "overdue" or "upcoming" for styling purposes.
func FormatDueDate(due *time.Time, status string) (string, string) {
	if due == nil {
		return "", "upcoming"
	}

	now := time.Now().In(time.Local)
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	taskDueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.Local)

	// If the due date is today (allowing a tiny tolerance), show remaining time until end of day
	if math.Abs(todayDate.Sub(taskDueDate).Hours()/24) < 0.01 {
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
		return fmt.Sprintf("%s (%d days overdue)", due.Format("2006-01-02"), daysOverdue), "overdue"
	} else {
		return due.Format("2006-01-02"), "upcoming"
	}
}
