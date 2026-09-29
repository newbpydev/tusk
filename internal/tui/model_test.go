package tui

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
)

func testOptions() Options {
	return Options{Context: context.Background(), Location: time.UTC, Profile: termenv.Ascii,
		Now:  func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) },
		Load: func(context.Context) ([]*core.TaskNode, error) { return nil, nil },
		Wait: func(ctx context.Context, _ time.Duration) error { return ctx.Err() }}
}

// Snapshot private widget fields as values, not pointer aliases. No warmup View
// is allowed: caches reached through the textarea are part of the contract.
func deepState(v reflect.Value, seen map[uintptr]bool) string {
	if !v.IsValid() {
		return "nil"
	}
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return "nil"
		}
		return deepState(v.Elem(), seen)
	case reflect.Pointer:
		if v.IsNil() {
			return "nil"
		}
		p := v.Pointer()
		if seen[p] {
			return fmt.Sprint(p)
		}
		seen[p] = true
		return fmt.Sprint(p) + deepState(v.Elem(), seen)
	case reflect.Struct:
		var b strings.Builder
		for i := 0; i < v.NumField(); i++ {
			b.WriteString(v.Type().Field(i).Name)
			b.WriteString(deepState(v.Field(i), seen))
		}
		return b.String()
	case reflect.Slice, reflect.Array:
		var b strings.Builder
		for i := 0; i < v.Len(); i++ {
			b.WriteString(deepState(v.Index(i), seen))
		}
		return b.String()
	case reflect.Map:
		var parts []string
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return deepState(keys[i], map[uintptr]bool{}) < deepState(keys[j], map[uintptr]bool{})
		})
		for _, k := range keys {
			parts = append(parts, deepState(k, seen)+":"+deepState(v.MapIndex(k), seen))
		}
		sort.Strings(parts)
		return strings.Join(parts, ";")
	case reflect.String:
		return fmt.Sprintf("%q", v.String())
	case reflect.Bool:
		return fmt.Sprint(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprint(v.Uint())
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return fmt.Sprint(v.Pointer())
	default:
		return v.Kind().String()
	}
}
func snapshot(v any) string { return deepState(reflect.ValueOf(v), map[uintptr]bool{}) }

func TestView_ColdAndRepeatedCallsPreserveDeepState(t *testing.T) {
	for _, populated := range []bool{false, true} {
		t.Run(fmt.Sprint(populated), func(t *testing.T) {
			m := New(testOptions())
			if populated {
				m.notes.SetValue("one\n界 two\nthree")
				m.notes.Focus()
				m.prepareFrame()
			}
			before := snapshot(m)
			frame := m.View()
			if frame == "" {
				t.Fatal("empty cold frame")
			}
			for range 100 {
				if m.View() != frame {
					t.Fatal("View changed output")
				}
			}
			if snapshot(m) != before {
				t.Fatal("View mutated deep reachable state")
			}
			m.notes.SetValue("changed")
			if snapshot(m) == before {
				t.Fatal("purity guard failed to detect nested mutation")
			}
		})
	}
}

func TestCommand_NoLiveModelCapture(t *testing.T) {
	o := testOptions()
	calls := 0
	o.Load = func(context.Context) ([]*core.TaskNode, error) { calls++; return nil, nil }
	o.Now = func() time.Time { t.Fatal("unexpected clock use"); return time.Time{} }
	o.Wait = func(context.Context, time.Duration) error { t.Fatal("unexpected wait"); return nil }
	m := New(o)
	cmd := m.Init()
	if calls != 0 {
		t.Fatal("constructor/Init performed I/O")
	}
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) { t.Fatal("captured live options"); return nil, nil }
	m.operation = 20
	before := snapshot(m)
	reply := cmd().(forestMsg)
	if reply.operation != 1 || calls != 1 || snapshot(m) != before {
		t.Fatal("command captured or mutated model")
	}
	m.Update(reply)
	if m.state != loading {
		t.Fatal("obsolete reply changed state")
	}
}

func TestTUI_LoadStates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		nodes []*core.TaskNode
		err   error
		want  string
	}{
		{"empty", nil, nil, "No tasks"},
		{"loaded", []*core.TaskNode{{Task: core.Task{ID: "one", Title: "demo"}}}, nil, "1 task"},
		{"error", nil, errors.New("private path"), "Could not load tasks"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := testOptions()
			o.Load = func(context.Context) ([]*core.TaskNode, error) { return tc.nodes, tc.err }
			m := New(o)
			if !strings.Contains(m.View(), "Loading") {
				t.Fatal(m.View())
			}
			m.Update(m.Init()())
			if !strings.Contains(m.View(), tc.want) || strings.Contains(m.View(), "private path") {
				t.Fatal(m.View())
			}
			m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			if m.width != 80 || m.height != 24 {
				t.Fatal("resize ignored")
			}
			before := m.View()
			m.Update(struct{}{})
			if m.View() != before {
				t.Fatal("unknown message changed frame")
			}
		})
	}
}

func TestTUI_SessionIsolation(t *testing.T) {
	o := testOptions()
	a := New(o)
	o.Profile = termenv.ANSI
	o.Location = time.FixedZone("test", 3600)
	b := New(o)
	if a.renderer == b.renderer || a.options.Location == b.options.Location {
		t.Fatal("shared session state")
	}
	if a.renderer.ColorProfile() != termenv.Ascii || b.renderer.ColorProfile() != termenv.ANSI {
		t.Fatal("profile leaked")
	}
	if strings.Contains(a.View(), "\x1b") {
		t.Fatal("plain frame contains escapes")
	}
}

func TestCommand_CanceledWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := Wait(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
}

func TestTUI_QuitAndRetry(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune{'q'}}, {Type: tea.KeyCtrlC}} {
		m := New(testOptions())
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatal("quit not dispatched")
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatal("not quit")
		}
		if (m.exitErr != nil) != (key.Type == tea.KeyCtrlC) {
			t.Fatal("incorrect exit outcome")
		}
	}
	m := New(testOptions())
	m.Update(forestMsg{operation: 1, err: errors.New("private")})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil || m.state != loading {
		t.Fatal("no retry")
	}
	m.Update(cmd())
	if m.state != loaded {
		t.Fatal("retry failed")
	}
}

func TestTUI_ServicePanicTerminatesSession(t *testing.T) {
	m := New(testOptions())
	_, cmd := m.Update(forestMsg{operation: 1, err: errRuntime})
	if cmd == nil || m.exitErr != errRuntime {
		t.Fatal("service panic left unsafe session open")
	}
}
