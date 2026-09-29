package main

import (
	"context"
	"os"
)

func processContext() (context.Context, func()) {
	return interruptContext(os.Interrupt)
}
