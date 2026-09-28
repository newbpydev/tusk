//go:build unix

package main

import (
	"context"
	"errors"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func confirm(ctx context.Context) (bool, error) { return confirmFD(ctx, int(os.Stdin.Fd())) }
func confirmFD(ctx context.Context, fd int) (bool, error) {
	return readConfirmation(ctx, func() (byte, error) {
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		for {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			n, err := unix.Poll(fds, 50)
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if err != nil {
				return 0, err
			}
			if n == 0 {
				continue
			}
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			if fds[0].Revents&unix.POLLNVAL != 0 {
				return 0, unix.EBADF
			}
			var b [1]byte
			n, err = unix.Read(fd, b[:])
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return 0, io.EOF
			}
			return b[0], nil
		}
	})
}
