package main

import (
	"context"
	"io"
	"os"

	"golang.org/x/sys/windows"
)

func confirm(ctx context.Context) (bool, error) {
	cancelIO := windows.NewLazySystemDLL("kernel32.dll").NewProc("CancelSynchronousIo")
	// Resolve support before any synchronous read starts.
	if err := cancelIO.Find(); err != nil {
		return false, err
	}
	return threadConfirmation(ctx, func() (func() (byte, error), func() error, func(), error) {
		thread, err := windows.OpenThread(windows.THREAD_TERMINATE, false, windows.GetCurrentThreadId())
		if err != nil {
			return nil, nil, nil, err
		}
		read := func() (byte, error) {
			var b [1]byte
			var n uint32
			err := windows.ReadFile(windows.Handle(os.Stdin.Fd()), b[:], &n, nil)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return 0, io.EOF
			}
			return b[0], nil
		}
		cancel := func() error {
			ok, _, err := cancelIO.Call(uintptr(thread))
			if ok == 0 {
				return err
			}
			return nil
		}
		return read, cancel, func() { _ = windows.CloseHandle(thread) }, nil
	})
}
