package main

import (
	"context"
	"os"
	"os/signal"
)

func processContext() (context.Context, func()) {
	return signal.NotifyContext(context.Background(), os.Interrupt)
}
