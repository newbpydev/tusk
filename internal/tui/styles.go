package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

const (
	canvasColor    = "#101820"
	dialogColor    = "#18232C"
	footerColor    = "#203039"
	selectionColor = "#193C42"
	textColor      = "#DCE7ED"
	mutedColor     = "#A1B1BE"
	borderColor    = "#657985"
	accentColor    = "#80DEC8"
	warningColor   = "#F7CC8C"
	dangerColor    = "#F2A99B"
)

func (m *Model) checkbox(label string, checked, focused bool) string {
	mark := "[ ] "
	if checked {
		mark = "[✓] "
	}
	if focused {
		return m.paint(">"+mark+label, accentColor, true)
	}
	return m.paint(" "+mark+label, textColor, false)
}

func (m *Model) actionButton(label string, focused, primary, disabled bool) string {
	color := textColor
	if primary || focused {
		color = accentColor
	}
	if disabled {
		color = mutedColor
		label += " · disabled"
	}
	text := "  [ " + label + " ]"
	if focused {
		text = "> [ " + label + " ]"
	}
	return m.paint(text, color, primary || focused)
}

func statusText(status core.Status) (string, string) {
	switch status {
	case core.StatusTodo:
		return "○", "Todo"
	case core.StatusInProgress:
		return "◐", "In progress"
	case core.StatusBlocked:
		return "!", "Blocked"
	case core.StatusDone:
		return "✓", "Done"
	default:
		return "○", terminaltext.Scalar(string(status))
	}
}

func priorityColor(priority core.Priority) string {
	if priority == core.PriorityUrgent {
		return dangerColor
	}
	if priority == core.PriorityHigh {
		return warningColor
	}
	return mutedColor
}

func priorityLabel(priority core.Priority) string {
	value := priority.String()
	if value != "" {
		return strings.ToUpper(value[:1]) + value[1:]
	}
	return value
}

func (m *Model) paint(s, color string, bold bool) string {
	return m.renderer.NewStyle().Foreground(lipgloss.Color(color)).Bold(bold).Render(s)
}

// surface wraps trusted, already-sanitized presentation. Child styles reset all
// SGR attributes when they end; restore the parent surface after these resets.
// Every boundary resets again so selected rows and overlays cannot bleed.
func (m *Model) surface(s, fg, bg string) string {
	if m.options.Profile == termenv.Ascii {
		return ansi.Strip(s)
	}
	p := m.options.Profile
	f, b := p.Convert(termenv.RGBColor(fg)), p.Convert(termenv.RGBColor(bg))
	base := "\x1b[" + f.Sequence(false) + ";" + b.Sequence(true) + "m"
	reset := "\x1b[0m"
	s = strings.NewReplacer(reset, reset+base, "\x1b[m", reset+base).Replace(s)
	return reset + base + s + reset
}
