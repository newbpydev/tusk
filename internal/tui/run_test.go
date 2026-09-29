package tui

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type failedOutput struct{ short bool }

func (w failedOutput) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, io.ErrClosedPipe
}

func TestSession_Run(t *testing.T) {
	for _, mode := range []string{"quit", "startup", "ctrlc", "panic", "update panic"} {
		t.Run(mode, func(t *testing.T) {
			opened, closed := 0, 0
			o := RunOptions{Input: strings.NewReader(""), Output: io.Discard, Location: time.UTC,
				Open: func(context.Context) (ports.TaskService, func() error, error) {
					opened++
					return sessionService{}, func() error { closed++; return nil }, nil
				},
				Program: func(m tea.Model, _ ...tea.ProgramOption) (tea.Model, error) {
					if mode == "startup" {
						return m, io.ErrClosedPipe
					}
					m.Update(m.Init()())
					if mode == "panic" {
						panic("PRIVATE")
					}
					if mode == "update panic" {
						m.(*Model).notes = textarea.Model{}
						m.(*Model).notes.Focus()
						m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
						return m, nil
					}
					key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
					if mode == "ctrlc" {
						key = tea.KeyMsg{Type: tea.KeyCtrlC}
					}
					m.Update(key)
					return m, nil
				}}
			_, err := Run(context.Background(), o)
			if (err == nil) != (mode == "quit") {
				t.Fatalf("error %v", err)
			}
			if closed != opened || (mode != "startup" && closed != 1) {
				t.Fatalf("open=%d close=%d", opened, closed)
			}
			if err != nil && strings.Contains(err.Error(), "PRIVATE") {
				t.Fatal("panic leaked")
			}
		})
	}
}

func TestSession_RecordingWriter(t *testing.T) {
	for _, short := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		w := &recordingWriter{writer: failedOutput{short}, cancel: cancel}
		_, err := w.Write([]byte("screen"))
		if err == nil || ctx.Err() == nil || w.failure() == nil {
			t.Fatal("failure not retained")
		}
		cancel()
	}
	var out bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := &recordingWriter{writer: &out, cancel: cancel}
	if _, err := w.Write([]byte("screen")); err != nil || w.failure() != nil || ctx.Err() != nil {
		t.Fatal("valid output failed")
	}
}

func TestSession_CommandPanic(t *testing.T) {
	o := testOptions()
	o.Load = func(context.Context) ([]*core.TaskNode, error) { panic("PRIVATE") }
	m := New(o)
	msg := m.Init()()
	m.Update(msg)
	if !errors.Is(m.exitErr, errRuntime) {
		t.Fatal("command panic not safe")
	}
}

func TestSession_UnknownCannotExitSuccessfully(t *testing.T) {
	unknown := ports.NewTransactionError("read", ports.ErrStorage)
	_, err := Run(context.Background(), RunOptions{Output: io.Discard, Open: func(context.Context) (ports.TaskService, func() error, error) { return nil, nil, unknown }, Program: func(m tea.Model, _ ...tea.ProgramOption) (tea.Model, error) { m.Update(m.Init()()); return m, nil }})
	if err == nil {
		t.Fatal("unknown outcome reported as clean exit")
	}
}

func TestSession_RealProgramAndOutputFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "quit", true: "output failure"}[fail], func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			var output bytes.Buffer
			var writer io.Writer = &output
			if fail {
				writer = failedOutput{}
			}
			_, err := Run(ctx, RunOptions{Input: strings.NewReader("q"), Output: writer, Open: func(context.Context) (ports.TaskService, func() error, error) {
				return sessionService{}, func() error { return nil }, nil
			}})
			if fail && !errors.Is(err, io.ErrClosedPipe) {
				t.Fatalf("lost writer error: %v", err)
			}
			if !fail && err != nil {
				t.Fatal(err)
			}
		})
	}
}
