package tui

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

type RunOptions struct {
	Input           io.Reader
	Output          io.Writer
	Open            Factory
	Location        *time.Location
	Profile         termenv.Profile
	Program         func(tea.Model, ...tea.ProgramOption) (tea.Model, error)
	CleanupProgress func()
}

func Run(ctx context.Context, options RunOptions) (result Result, err error) {
	ctx, cancel := context.WithCancel(ctx)
	session := NewSession(ctx, options.Open)
	out := &recordingWriter{writer: options.Output, cancel: cancel}
	defer func() {
		if recover() != nil {
			err = errRuntime
		}
		cancel()
		var closeErr error
		result, closeErr = session.Close(options.CleanupProgress)
		err = errors.Join(err, out.failure(), closeErr)
		if result.OutcomeUnknown && err == nil {
			err = errRuntime
		}
	}()
	m := New(Options{Context: ctx, Load: session.Load, Now: time.Now, Wait: Wait, Location: options.Location, Profile: options.Profile})
	program := options.Program
	if program == nil {
		program = func(m tea.Model, opts ...tea.ProgramOption) (tea.Model, error) {
			return tea.NewProgram(m, opts...).Run()
		}
	}
	_, err = program(m, tea.WithInput(options.Input), tea.WithOutput(out), tea.WithContext(ctx), tea.WithAltScreen(), tea.WithoutSignalHandler())
	return result, errors.Join(err, m.exitErr)
}

type recordingWriter struct {
	writer io.Writer
	cancel context.CancelFunc
	mu     sync.Mutex
	err    error
}

func (w *recordingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	if err != nil && w.err == nil {
		w.err = err
		w.cancel()
	}
	return n, err
}
func (w *recordingWriter) failure() error { w.mu.Lock(); defer w.mu.Unlock(); return w.err }

type fatalMsg struct{ err error }

func safeCommand(cmd tea.Cmd) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if recover() != nil {
				msg = fatalMsg{errRuntime}
			}
		}()
		return cmd()
	}
}
