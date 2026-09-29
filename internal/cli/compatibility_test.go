package cli

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/termenv"
)

func TestCompatibility_Presentation(t *testing.T) {
	var out bytes.Buffer
	renderer := lipgloss.NewRenderer(&out, termenv.WithProfile(termenv.Ascii))
	renderer.SetColorProfile(termenv.Ascii)
	renderer.SetHasDarkBackground(true)
	if got := renderer.NewStyle().Foreground(lipgloss.Color("2")).Render("tusk"); got != "tusk" {
		t.Fatal(got)
	}
	if ansi.StringWidth("界") != 2 {
		t.Fatal("cell width")
	}
	// Reference the platform API without making a terminal probe.
	_ = term.GetSize
	if out.Len() != 0 {
		t.Fatalf("terminal traffic %q", out.String())
	}
}
