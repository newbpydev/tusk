package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
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
)

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
