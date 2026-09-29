package main

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/newbpydev/tusk/internal/ports"
)

func readConfirmation(ctx context.Context, readByte func() (byte, error)) (bool, error) {
	line := make([]byte, 0, 32)
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		b, err := readByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			return false, err
		}
		if err = ctx.Err(); err != nil {
			return false, err
		}
		if b == '\n' {
			answer := strings.TrimSpace(string(line))
			return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes"), nil
		}
		if len(line) >= 4096 {
			return false, ports.ErrInvalidText
		}
		line = append(line, b)
	}
}
