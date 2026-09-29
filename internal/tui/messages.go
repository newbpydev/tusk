package tui

import (
	"github.com/newbpydev/tusk/internal/core"
	"time"
)

type forestMsg struct {
	operation         uint64
	owner, generation uint64
	now               time.Time
	forest            []*core.TaskNode
	err               error
}
