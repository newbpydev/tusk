//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestConfirm_UnixPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if _, err = w.WriteString("yes\r\n"); err != nil {
		t.Fatal(err)
	}
	if yes, err := confirmFD(context.Background(), int(r.Fd())); !yes || err != nil {
		t.Fatalf("%v %v", yes, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	start := time.Now()
	if yes, err := confirmFD(ctx, int(r.Fd())); yes || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("%v %v", yes, err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("unbounded poll")
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if yes, err := confirmFD(context.Background(), int(r.Fd())); yes || err != nil {
		t.Fatalf("EOF %v %v", yes, err)
	}
	fd := int(r.Fd())
	if err = r.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = confirmFD(context.Background(), fd); err == nil {
		t.Fatal("closed descriptor accepted")
	}
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	if _, err = confirm(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestConfirm_UnixReadError(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if err = r.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if yes, err := confirmFD(ctx, int(w.Fd())); yes || err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("write-only descriptor: %v %v", yes, err)
	}
}
