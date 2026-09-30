package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// These are civil dates represented in UTC, not due instants. Navigation must
// not skip or repeat days when the configured location changes its UTC offset.
type datePicker struct {
	selected, today time.Time
}

func civilDate(t time.Time) time.Time {
	y, month, day := t.Date()
	return time.Date(y, month, day, 0, 0, 0, 0, time.UTC)
}

func (m *Model) beginCalendar() {
	now := m.options.Now()
	today := civilDate(now.In(m.options.Location))
	selected := today
	if parse := m.options.ParseDue; parse != nil {
		if due, err := parse(m.form.draft.fields[fieldDue], now, m.options.Location); err == nil {
			selected = civilDate(due.In(m.options.Location))
		}
	}
	m.form.calendar = &datePicker{selected: selected, today: today}
	m.form.err = ""
}

func (m *Model) calendarKey(key tea.KeyMsg) {
	f := m.form
	p := f.calendar
	next := p.selected
	switch key.String() {
	case "esc":
		f.calendar = nil
		return
	case "enter":
		value := p.selected.Format("2006-01-02")
		f.draft.fields[fieldDue] = value
		f.inputs[fieldDue].SetValue(value)
		f.calendar = nil
		return
	case "left":
		next = next.AddDate(0, 0, -1)
	case "right":
		next = next.AddDate(0, 0, 1)
	case "up":
		next = next.AddDate(0, 0, -7)
	case "down":
		next = next.AddDate(0, 0, 7)
	case "pgup", "pgdown":
		delta := 1
		if key.String() == "pgup" {
			delta = -1
		}
		first := time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, delta, 0)
		last := first.AddDate(0, 1, -1).Day()
		next = first.AddDate(0, 0, min(next.Day(), last)-1)
	case "t":
		p.today = civilDate(m.options.Now().In(m.options.Location))
		next = p.today
	}
	if next.Year() >= 1 && next.Year() <= 9999 {
		p.selected = next
	}
}

func (m *Model) calendarLines(width int) []string {
	p := m.form.calendar
	center := func(s string) string {
		return strings.Repeat(" ", max(0, (width-ansi.StringWidth(s))/2)) + s
	}
	first := time.Date(p.selected.Year(), p.selected.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := (int(first.Weekday()) + 6) % 7 // Monday first.
	last := first.AddDate(0, 1, -1).Day()
	lines := []string{
		center(m.paint("‹  "+p.selected.Format("January 2006")+"  ›", accentColor, true)), "",
		center(m.paint(" Mon   Tue   Wed   Thu   Fri   Sat   Sun  ", mutedColor, false)),
	}
	for week := 0; week < 6; week++ {
		var row strings.Builder
		for weekday := 0; weekday < 7; weekday++ {
			day := week*7 + weekday - start + 1
			if day < 1 || day > last {
				row.WriteString("      ")
				continue
			}
			mark := " "
			if first.Year() == p.today.Year() && first.Month() == p.today.Month() && day == p.today.Day() {
				mark = "•"
			}
			cell := fmt.Sprintf(" %2d%s  ", day, mark)
			if day == p.selected.Day() {
				cell = m.surface(m.paint(fmt.Sprintf("[%2d]%s ", day, mark), accentColor, true), accentColor, selectionColor)
			}
			row.WriteString(cell)
		}
		lines = append(lines, center(row.String()))
	}
	return append(lines, "",
		center(m.paint(p.selected.Format("Mon, 2 Jan 2006")+" · "+p.selected.Format("2006-01-02"), textColor, true)),
		center(m.paint("• Today · "+m.options.Location.String(), mutedColor, false)))
}
