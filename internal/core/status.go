package core

import (
	"strings"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

func ParseStatus(s string) (Status, error) {
	norm := strings.ToLower(strings.TrimSpace(s))
	switch Status(norm) {
	case StatusTodo:
		return StatusTodo, nil
	case StatusInProgress:
		return StatusInProgress, nil
	case StatusBlocked:
		return StatusBlocked, nil
	case StatusDone:
		return StatusDone, nil
	default:
		return "", ErrInvalidStatus
	}
}

func (s Status) IsValid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusBlocked, StatusDone:
		return true
	default:
		return false
	}
}

func (s Status) IsTerminal() bool {
	return s == StatusDone
}

func (s Status) CanTransitionTo(next Status) bool {
	if !s.IsValid() || !next.IsValid() {
		return false
	}
	if s == next {
		return true
	}
	switch s {
	case StatusTodo, StatusInProgress, StatusBlocked:
		return true
	case StatusDone:
		return next == StatusTodo || next == StatusInProgress
	default:
		return false
	}
}
