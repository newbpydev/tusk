package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

// Darkmatter dark preset from tweakcn. Borders use its accent gray for visible
// terminal strokes; destructive actions retain the preset's red light token.
const (
	canvasColor    = "#121113"
	dialogColor    = "#121113"
	footerColor    = "#222222"
	selectionColor = "#333333"
	textColor      = "#C1C1C1"
	mutedColor     = "#888888"
	borderColor    = "#333333"
	accentColor    = "#E78A53"
	warningColor   = "#FBCB97"
	dangerColor    = "#EF4444"
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
