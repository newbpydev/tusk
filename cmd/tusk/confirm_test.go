package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestConfirm_Lines(t *testing.T) {
	for _, tc := range []struct {
		line      string
		yes       bool
		wantError bool
	}{{"y\n", true, false}, {" YES \r\n", true, false}, {"\n", false, false}, {"no\n", false, false}, {"anything\n", false, false}, {"", false, false}, {"yes", false, false}, {strings.Repeat("a", 4097) + "\n", false, true}} {
		r := strings.NewReader(tc.line)
		yes, err := readConfirmation(context.Background(), func() (byte, error) { return r.ReadByte() })
		if yes != tc.yes || (err != nil) != tc.wantError {
			t.Fatalf("line len %d yes %v err %v", len(tc.line), yes, err)
		}
	}
}
func TestConfirm_ReadFailureAndCancel(t *testing.T) {
	if _, err := readConfirmation(context.Background(), func() (byte, error) { return 0, io.ErrClosedPipe }); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readConfirmation(ctx, func() (byte, error) { t.Fatal("read after cancel"); return 0, nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	r := strings.NewReader("yes\n")
	if yes, err := readConfirmation(ctx, func() (byte, error) {
		b, e := r.ReadByte()
		if b == '\n' {
			cancel()
		}
		return b, e
	}); yes || !errors.Is(err, context.Canceled) {
		t.Fatalf("%v %v", yes, err)
	}
}

func TestConfirm_CancelJoinsReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	unblock := make(chan struct{})
	released := make(chan struct{})
	closed := false
	attempts := 0
	yes, err := threadConfirmation(ctx, func() (func() (byte, error), func() error, func(), error) {
		read := func() (byte, error) { close(started); cancel(); <-unblock; <-released; return 0, io.ErrClosedPipe }
		stop := func() error {
			attempts++
			if attempts == 1 {
				return errors.New("no pending IO")
			}
			if attempts == 3 {
				close(released)
			}
			if attempts == 2 {
				close(unblock)
			}
			return nil
		}
		return read, stop, func() { closed = true }, nil
	})
	if yes || !errors.Is(err, context.Canceled) || !closed || attempts < 3 {
		t.Fatalf("yes %v err %v closed %v attempts %d", yes, err, closed, attempts)
	}
}
func TestConfirm_ThreadSetupFailure(t *testing.T) {
	yes, err := threadConfirmation(context.Background(), func() (func() (byte, error), func() error, func(), error) { return nil, nil, nil, io.ErrClosedPipe })
	if yes || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("%v %v", yes, err)
	}
}
func TestConfirm_ThreadSuccess(t *testing.T) {
	r := strings.NewReader("yes\n")
	closed := false
	yes, err := threadConfirmation(context.Background(), func() (func() (byte, error), func() error, func(), error) {
		return r.ReadByte, func() error { t.Error("unneeded cancel"); return nil }, func() { closed = true }, nil
	})
	if !yes || err != nil || !closed {
		t.Fatalf("%v %v %v", yes, err, closed)
	}
}
