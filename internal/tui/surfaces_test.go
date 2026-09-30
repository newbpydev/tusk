package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

// Terminal-cell background verification guards the shadow artifacts found in
// the Kitty prototype. A plain string/screenshot golden misses SGR inheritance.
func TestOverlay_UniformSurfaces(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 60}} {
		for _, modalKind := range []string{"none", "help", "form", "delete", "recovery"} {
			modal := modalKind != "none"
			t.Run(fmt.Sprintf("%dx%d/modal=%s", size[0], size[1], modalKind), func(t *testing.T) {
				o := testOptions()
				o.Profile = termenv.TrueColor
				m := New(o)
				m.state = loaded
				m.forest = []*core.TaskNode{{Task: core.Task{Title: "A task", Status: core.StatusTodo, Priority: core.PriorityMedium}}}
				m.helpOpen = modalKind == "help"
				if modalKind == "form" {
					m.beginForm(false)
					m.form.err = "Enter a title."
				}
				if modalKind == "delete" {
					m.confirmation = &deleteDialog{target: taskRef{id: "one"}, title: "A task", preview: &ports.DeletePreview{IDs: []string{"one", "child"}}}
				}
				if modalKind == "recovery" {
					m.freezeWrites()
				}
				m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
				l := measure(size[0], size[1])
				b := l.modal
				for y, row := range strings.Split(m.View(), "\n") {
					bg := "default"
					x := 0
					var state byte
					for len(row) > 0 {
						seq, width, n, next := ansi.DecodeSequence(row, state, nil)
						row = row[n:]
						state = next
						if strings.HasPrefix(seq, "\x1b[") && strings.HasSuffix(seq, "m") {
							params := strings.Split(seq[2:len(seq)-1], ";")
							for i := 0; i < len(params); i++ {
								switch params[i] {
								case "", "0", "49":
									bg = "default"
								case "38", "48":
									if i+4 < len(params) && params[i+1] == "2" {
										if params[i] == "48" {
											bg = strings.Join(params[i+2:i+5], ",")
										}
										i += 4
									}
								}
							}
						}
						for cell := 0; cell < width; cell++ {
							want := "16,24,32"
							if y == l.height-1 {
								want = "32,48,56"
							}
							if modal && x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height {
								want = "24,35,44"
							}
							if !modal && (y == 5 || y == 6) && x >= 1 && x < l.listWidth-1 {
								want = "25,60,65"
							}
							if bg != want {
								t.Fatalf("cell (%d,%d) background=%s want %s", x, y, bg, want)
							}
							x++
						}
					}
					if bg != "default" {
						t.Fatalf("row %d leaks background %s", y, bg)
					}
				}
			})
		}
	}
}
