// Package input provides input components for the TUI application
package input

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DateInputMode represents the current editing mode of the date input
type DateInputMode int

const (
	// DateModeEmpty is the initial mode when no date is entered
	DateModeEmpty DateInputMode = iota
	// DateModeView is the mode when date is populated but not actively editing
	DateModeView
	// DateModeDateEdit is the mode for editing the entire date part
	DateModeDateEdit
	// DateModeTimeEdit is the mode for editing the entire time part
	DateModeTimeEdit
	// DateModeYearEdit is the mode for editing the year component
	DateModeYearEdit
	// DateModeMonthEdit is the mode for editing the month component
	DateModeMonthEdit
	// DateModeDayEdit is the mode for editing the day component
	DateModeDayEdit
	// DateModeHourEdit is the mode for editing the hour component
	DateModeHourEdit
	// DateModeMinuteEdit is the mode for editing the minute component
	DateModeMinuteEdit
)

// Helper constants to define groups of modes
const (
	// DateGroup represents date editing modes
	DateGroupDate = 1
	// DateGroupTime represents time editing modes
	DateGroupTime = 2
)

// predefined time hours for quick selection
var timeHourOptions = []int{0, 2, 6, 8, 10, 12, 14, 16, 18, 20, 22}

// DateInput represents an interactive date input field
type DateInput struct {
	// Value holds the current date time value
	Value time.Time
	// HasValue indicates if a date has been entered
	HasValue bool
	// Mode represents the current editing mode
	Mode DateInputMode
	// Label is the field label
	Label string
	// Focused indicates if the input has focus
	Focused bool
	// Error holds any validation error
	Error string
	// BaseStyle holds the base styling for the input
	BaseStyle lipgloss.Style
	// FocusedStyle holds the styling when input is focused
	FocusedStyle lipgloss.Style
	// ErrorStyle holds the styling when input has an error
	ErrorStyle lipgloss.Style
}

// NewDateInput creates a new date input component
func NewDateInput(label string) *DateInput {
	return &DateInput{
		Label:       label,
		Mode:        DateModeEmpty,
		HasValue:    false,
		BaseStyle:   lipgloss.NewStyle().PaddingLeft(1),
		FocusedStyle: lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0D6EFD")),
		ErrorStyle: lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("#FF0000")),
	}
}

// StringValue returns the current date as formatted string with both date and time
func (d *DateInput) StringValue() string {
	if !d.HasValue {
		return ""
	}
	// Always include time in the output for consistency
	return d.Value.Format("2006-01-02 15:04")
}

// DateString returns just the date part
func (d *DateInput) DateString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("2006-01-02")
}

// TimeString returns just the time part
func (d *DateInput) TimeString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("15:04")
}

// YearString returns just the year component
func (d *DateInput) YearString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("2006")
}

// MonthString returns just the month component
func (d *DateInput) MonthString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("01")
}

// DayString returns just the day component
func (d *DateInput) DayString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("02")
}

// HourString returns just the hour component
func (d *DateInput) HourString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("15")
}

// MinuteString returns just the minute component
func (d *DateInput) MinuteString() string {
	if !d.HasValue {
		return ""
	}
	return d.Value.Format("04")
}

// SetToToday sets the date to today at 00:00
func (d *DateInput) SetToToday() {
	now := time.Now()
	// Set to today at 00:00
	d.Value = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d.HasValue = true
	d.Mode = DateModeView
}

// Reset clears the input
func (d *DateInput) Reset() {
	d.HasValue = false
	d.Mode = DateModeEmpty
	d.Error = ""
}

// IncrementDay adds one day to the current date
func (d *DateInput) IncrementDay() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	d.Value = d.Value.AddDate(0, 0, 1)
}

// DecrementDay subtracts one day from the current date
func (d *DateInput) DecrementDay() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	d.Value = d.Value.AddDate(0, 0, -1)
}

// IncrementHour cycles through predefined hour options
func (d *DateInput) IncrementHour() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	
	currentHour := d.Value.Hour()
	nextHour := 0
	
	for i, hour := range timeHourOptions {
		if currentHour < hour {
			nextHour = hour
			break
		}
		if i == len(timeHourOptions)-1 {
			// Wrap around to first option
			nextHour = timeHourOptions[0]
		}
	}
	
	d.Value = time.Date(
		d.Value.Year(),
		d.Value.Month(),
		d.Value.Day(),
		nextHour,
		d.Value.Minute(),
		0,
		0,
		d.Value.Location(),
	)
}

// DecrementHour cycles through predefined hour options in reverse
func (d *DateInput) DecrementHour() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	
	currentHour := d.Value.Hour()
	prevHour := 22 // Default to last hour option
	
	for i := len(timeHourOptions) - 1; i >= 0; i-- {
		hour := timeHourOptions[i]
		if currentHour > hour {
			prevHour = hour
			break
		}
		if i == 0 {
			// Wrap around to last option
			prevHour = timeHourOptions[len(timeHourOptions)-1]
		}
	}
	
	d.Value = time.Date(
		d.Value.Year(),
		d.Value.Month(),
		d.Value.Day(),
		prevHour,
		d.Value.Minute(),
		0,
		0,
		d.Value.Location(),
	)
}

// ComponentIncrement increments the currently focused component
func (d *DateInput) ComponentIncrement() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	
	switch d.Mode {
	case DateModeDateEdit:
		d.IncrementDay()
	case DateModeTimeEdit:
		d.IncrementHour()
	case DateModeYearEdit:
		d.Value = d.Value.AddDate(1, 0, 0)
	case DateModeMonthEdit:
		if d.Value.Month() == time.December {
			d.Value = time.Date(
				d.Value.Year(),
				time.January,
				d.Value.Day(),
				d.Value.Hour(),
				d.Value.Minute(),
				0,
				0,
				d.Value.Location(),
			)
		} else {
			d.Value = d.Value.AddDate(0, 1, 0)
		}
	case DateModeDayEdit:
		d.IncrementDay()
	case DateModeHourEdit:
		currentHour := d.Value.Hour()
		nextHour := (currentHour + 1) % 24
		d.Value = time.Date(
			d.Value.Year(),
			d.Value.Month(),
			d.Value.Day(),
			nextHour,
			d.Value.Minute(),
			0,
			0,
			d.Value.Location(),
		)
	case DateModeMinuteEdit:
		currentMinute := d.Value.Minute()
		nextMinute := (currentMinute + 5) % 60
		d.Value = time.Date(
			d.Value.Year(),
			d.Value.Month(),
			d.Value.Day(),
			d.Value.Hour(),
			nextMinute,
			0,
			0,
			d.Value.Location(),
		)
	}
}

// ComponentDecrement decrements the currently focused component
func (d *DateInput) ComponentDecrement() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	
	switch d.Mode {
	case DateModeDateEdit:
		d.DecrementDay()
	case DateModeTimeEdit:
		d.DecrementHour()
	case DateModeYearEdit:
		d.Value = d.Value.AddDate(-1, 0, 0)
	case DateModeMonthEdit:
		if d.Value.Month() == time.January {
			d.Value = time.Date(
				d.Value.Year(),
				time.December,
				d.Value.Day(),
				d.Value.Hour(),
				d.Value.Minute(),
				0,
				0,
				d.Value.Location(),
			)
		} else {
			d.Value = d.Value.AddDate(0, -1, 0)
		}
	case DateModeDayEdit:
		d.DecrementDay()
	case DateModeHourEdit:
		currentHour := d.Value.Hour()
		nextHour := (currentHour - 1 + 24) % 24
		d.Value = time.Date(
			d.Value.Year(),
			d.Value.Month(),
			d.Value.Day(),
			nextHour,
			d.Value.Minute(),
			0,
			0,
			d.Value.Location(),
		)
	case DateModeMinuteEdit:
		currentMinute := d.Value.Minute()
		nextMinute := (currentMinute - 5 + 60) % 60
		d.Value = time.Date(
			d.Value.Year(),
			d.Value.Month(),
			d.Value.Day(),
			d.Value.Hour(),
			nextMinute,
			0,
			0,
			d.Value.Location(),
		)
	}
}

// EnterNextMode advances to the next editing mode based on the current section
func (d *DateInput) EnterNextMode() {
	if !d.HasValue {
		d.SetToToday()
		return
	}
	
	switch d.Mode {
	case DateModeEmpty, DateModeView:
		// Start with date editing by default
		d.Mode = DateModeDateEdit
		
	case DateModeDateEdit:
		// When on date section and pressing Enter, go directly to year edit
		d.Mode = DateModeYearEdit
		
	case DateModeTimeEdit:
		// When on time section and pressing Enter, go directly to hour edit
		d.Mode = DateModeHourEdit
		
	// Date component progression
	case DateModeYearEdit:
		d.Mode = DateModeMonthEdit
	case DateModeMonthEdit:
		d.Mode = DateModeDayEdit
	case DateModeDayEdit:
		// When done with day, return to date edit mode rather than going to time
		d.Mode = DateModeDateEdit
		
	// Time component progression
	case DateModeHourEdit:
		d.Mode = DateModeMinuteEdit
	case DateModeMinuteEdit:
		// When done with minute, return to time edit mode rather than continuing
		d.Mode = DateModeTimeEdit
	}
}

// HandleInput processes keyboard input for the date input
func (d *DateInput) HandleInput(msg tea.KeyMsg) {
	// First handle special keys that should always work in any mode
	switch msg.Type {
	case tea.KeyEsc:
		// Handle Esc key with multiple levels
		if d.HasValue {
			// If in a deep editing mode (year/month/day or hour/minute), go up one level
			if d.Mode >= DateModeYearEdit && d.Mode <= DateModeDayEdit {
				// From date component editing, return to date selection
				d.Mode = DateModeDateEdit
				return
			} else if d.Mode >= DateModeHourEdit && d.Mode <= DateModeMinuteEdit {
				// From time component editing, return to time selection
				d.Mode = DateModeTimeEdit
				return
			} else if d.Mode == DateModeDateEdit || d.Mode == DateModeTimeEdit {
				// From date/time selection, return to view mode
				d.Mode = DateModeView
				return
			}
			// Otherwise (in view mode), let the form handle the Esc key
		}
		return
	case tea.KeyTab, tea.KeyShiftTab:
		// Tab navigation is handled at a higher level, but we should make sure
		// we're not in edit mode when tabbing away
		if d.HasValue {
			d.Mode = DateModeView
		}
		return
	case tea.KeyLeft, tea.KeyRight:
		// Handle left/right navigation between date and time sections
		if d.HasValue {
			d.handleLeftRightNavigation(msg.Type)
			return
		}
	}

	// Then handle specific keys for different operations
	switch msg.String() {
	case " ":
		if !d.HasValue {
			d.SetToToday()
		}
	case "enter":
		d.EnterNextMode()
	case "up":
		d.ComponentIncrement()
	case "down":
		d.ComponentDecrement()
	case "backspace":
		// Handle backspace to clear field
		if d.HasValue && d.Mode == DateModeView {
			d.Reset()
		}
	}
}

// SetValue sets the input value from a time.Time
func (d *DateInput) SetValue(t time.Time) {
	d.Value = t
	d.HasValue = true
	d.Mode = DateModeView
}

// SetValueFromString sets the input value from a string in format YYYY-MM-DD HH:MM
func (d *DateInput) SetValueFromString(dateStr string) error {
	if dateStr == "" {
		d.Reset()
		return nil
	}
	
	// Try to parse full datetime string
	t, err := time.Parse("2006-01-02 15:04", dateStr)
	if err != nil {
		// Try date only format
		t, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
	}
	
	d.SetValue(t)
	return nil
}

// getEditingGroup returns which group (date or time) the current mode belongs to
func (d *DateInput) getEditingGroup() int {
	switch d.Mode {
	case DateModeDateEdit, DateModeYearEdit, DateModeMonthEdit, DateModeDayEdit:
		return DateGroupDate
	case DateModeTimeEdit, DateModeHourEdit, DateModeMinuteEdit:
		return DateGroupTime
	default:
		return 0 // Not in an editing group
	}
}

// handleLeftRightNavigation handles all left/right arrow navigation
func (d *DateInput) handleLeftRightNavigation(keyType tea.KeyType) {
	// Only handle left/right in view mode or when already in an edit mode
	if d.Mode == DateModeEmpty {
		return
	}

	// First handle the base view mode - this lets users select date or time
	if d.Mode == DateModeView {
		if keyType == tea.KeyLeft {
			// Left arrow from view mode goes to date editing
			d.Mode = DateModeDateEdit
		} else if keyType == tea.KeyRight {
			// Right arrow from view mode goes to time editing
			d.Mode = DateModeTimeEdit
		}
		return
	}

	// Then handle the top-level editing modes (date vs time)
	if d.Mode == DateModeDateEdit || d.Mode == DateModeTimeEdit {
		if keyType == tea.KeyLeft && d.Mode == DateModeTimeEdit {
			// Switch from time to date editing
			d.Mode = DateModeDateEdit
			return
		} else if keyType == tea.KeyRight && d.Mode == DateModeDateEdit {
			// Switch from date to time editing
			d.Mode = DateModeTimeEdit
			return
		}
	}

	// Then handle the detailed date component navigation
	if d.Mode >= DateModeYearEdit && d.Mode <= DateModeDayEdit {
		if keyType == tea.KeyLeft {
			// Move to previous date component
			switch d.Mode {
			case DateModeMonthEdit:
				d.Mode = DateModeYearEdit
			case DateModeDayEdit:
				d.Mode = DateModeMonthEdit
			}
		} else if keyType == tea.KeyRight {
			// Move to next date component
			switch d.Mode {
			case DateModeYearEdit:
				d.Mode = DateModeMonthEdit
			case DateModeMonthEdit:
				d.Mode = DateModeDayEdit
			}
		}
		return
	}

	// Finally handle time component navigation
	if d.Mode >= DateModeHourEdit && d.Mode <= DateModeMinuteEdit {
		if keyType == tea.KeyLeft {
			// Move to hour when on minute
			if d.Mode == DateModeMinuteEdit {
				d.Mode = DateModeHourEdit
			}
		} else if keyType == tea.KeyRight {
			// Move to minute when on hour
			if d.Mode == DateModeHourEdit {
				d.Mode = DateModeMinuteEdit
			}
		}
	}
}

// View renders the date input component
func (d *DateInput) View() string {
	var content string
	
	if d.HasValue {
		// Format with a consistent format that always shows date and time
		switch d.Mode {
		case DateModeView:
			// In view mode, show date and time without brackets, simple dash like priority field
			formattedDate := d.DateString()
			formattedTime := d.TimeString()
			content = fmt.Sprintf("%s %s - ← → to select", formattedDate, formattedTime)
		case DateModeDateEdit:
			content = fmt.Sprintf("[%s] %s - Enter to edit date components", d.DateString(), d.TimeString())
		case DateModeTimeEdit:
			content = fmt.Sprintf("%s [%s] - Enter to edit time components", d.DateString(), d.TimeString())
		case DateModeYearEdit:
			content = fmt.Sprintf("[%s]-%s-%s %s - ← → to navigate", d.YearString(), d.MonthString(), d.DayString(), d.TimeString())
		case DateModeMonthEdit:
			content = fmt.Sprintf("%s-[%s]-%s %s - ← → to navigate", d.YearString(), d.MonthString(), d.DayString(), d.TimeString())
		case DateModeDayEdit:
			content = fmt.Sprintf("%s-%s-[%s] %s - ← → to navigate", d.YearString(), d.MonthString(), d.DayString(), d.TimeString())
		case DateModeHourEdit:
			content = fmt.Sprintf("%s [%s]:%s - ← → to navigate", d.DateString(), d.HourString(), d.MinuteString())
		case DateModeMinuteEdit:
			content = fmt.Sprintf("%s %s:[%s] - ← → to navigate", d.DateString(), d.HourString(), d.MinuteString())
		}
	} else {
		content = "Optional - Space to set today's date"
	}
	
	style := d.BaseStyle
	if d.Focused {
		style = d.FocusedStyle
	}
	if d.Error != "" {
		style = d.ErrorStyle
	}
	
	// Render the label and content with highlighted label when focused
	var field string
	if d.Focused {
		// Use blue text for the label when focused to match other fields
		labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#0D6EFD"))
		field = fmt.Sprintf("%s: %s", labelStyle.Render(d.Label), style.Render(content))
	} else {
		labelStyle := lipgloss.NewStyle().Bold(true)
		field = fmt.Sprintf("%s: %s", labelStyle.Render(d.Label), style.Render(content))
	}
	
	// Add error message if present
	if d.Error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		field += "\n" + errorStyle.Render(d.Error)
	}
	
	return field
}
