package core

import (
	"strconv"
	"strings"
)

type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
	PriorityUrgent Priority = 4
)

func ParsePriority(s string) (Priority, error) {
	norm := strings.ToLower(strings.TrimSpace(s))
	switch norm {
	case "low", "1":
		return PriorityLow, nil
	case "medium", "2":
		return PriorityMedium, nil
	case "high", "3":
		return PriorityHigh, nil
	case "urgent", "4":
		return PriorityUrgent, nil
	}

	if n, err := strconv.Atoi(norm); err == nil {
		p := Priority(n)
		if p.IsValid() {
			return p, nil
		}
	}

	return 0, ErrInvalidPriority
}

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return ""
	}
}

func (p Priority) Weight() int {
	return int(p)
}

func (p Priority) IsValid() bool {
	return p >= PriorityLow && p <= PriorityUrgent
}
