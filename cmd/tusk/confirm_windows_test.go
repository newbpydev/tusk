package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestConfirm_WindowsPipe(t *testing.T) {
	for _, tc := range []struct {
		line string
		yes  bool
	}{{"yes\r\n", true}, {" no \r\n", false}, {"", false}, {"yes", false}} {
		t.Run(tc.line, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			defer w.Close()
			if _, err := w.WriteString(tc.line); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			if yes, err := confirmHandle(context.Background(), windows.Handle(r.Fd())); yes != tc.yes || err != nil {
				t.Fatalf("answer %v, error %v", yes, err)
			}
		})
	}
}

func TestConfirm_WindowsFileAndInvalidHandle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty-input")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if yes, err := confirmHandle(context.Background(), windows.Handle(f.Fd())); yes || err != nil {
		t.Fatalf("file EOF: %v %v", yes, err)
	}
	if yes, err := confirmHandle(context.Background(), windows.InvalidHandle); yes || !errors.Is(err, windows.ERROR_INVALID_HANDLE) {
		t.Fatalf("invalid handle: %v %v", yes, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if yes, err := confirm(ctx); yes || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel before reading stdin: %v %v", yes, err)
	}
}

func TestConfirm_WindowsCancelJoinsNativeReader(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	type result struct {
		yes bool
		err error
	}
	done := make(chan result, 1)
	go func() {
		yes, err := confirmHandle(ctx, windows.Handle(r.Fd()))
		done <- result{yes, err}
	}()
	select {
	case answer := <-done:
		if answer.yes || !errors.Is(answer.err, context.DeadlineExceeded) {
			t.Fatalf("cancel: %v %v", answer.yes, answer.err)
		}
	case <-time.After(2 * time.Second):
		// Release the owned writer and join before reporting failed OS cancellation.
		_ = w.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("native reader did not join after owned pipe closure")
		}
		t.Fatal("native cancellation did not release the reader")
	}
}
