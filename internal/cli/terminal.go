package cli

import (
	"io"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TerminalFacts are sampled only when a data command needs terminal behavior.
type TerminalFacts struct {
	In, Out, Err bool
	Width        int
	WidthError   error
}
type formatter struct {
	terminal bool
	width    int
	location *time.Location
	renderer *lipgloss.Renderer
}

func newFormatter(facts TerminalFacts, location *time.Location, getenv func(string) string) *formatter {
	if location == nil {
		location = time.Local
	}
	width := facts.Width
	if facts.WidthError != nil {
		width = 80
	}
	if width < 1 {
		width = 1
	}
	f := &formatter{terminal: facts.Out, width: width, location: location}
	if facts.Out {
		profile := termenv.ANSI
		if getenv != nil && (getenv("NO_COLOR") != "" || getenv("TERM") == "dumb") {
			profile = termenv.Ascii
		}
		f.renderer = lipgloss.NewRenderer(io.Discard, termenv.WithProfile(profile))
		f.renderer.SetColorProfile(profile)
		f.renderer.SetHasDarkBackground(true)
	}
	return f
}

func (i *invocation) terminal() TerminalFacts {
	if !i.terminalSampled {
		i.terminalSampled = true
		if i.options.Terminal != nil {
			i.terminalFacts = i.options.Terminal()
		}
	}
	return i.terminalFacts
}
func (i *invocation) formatResult(value any, jsonOutput bool) ([]byte, error) {
	if jsonOutput {
		return encodeJSON(value)
	}
	return newFormatter(i.terminal(), i.configValue.Location, i.options.Getenv).format(value)
}
