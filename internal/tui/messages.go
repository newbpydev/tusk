package tui

import "github.com/newbpydev/tusk/internal/core"

type forestMsg struct {
	operation uint64
	forest    []*core.TaskNode
	err       error
}
