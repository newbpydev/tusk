package core

import (
	"sort"
	"strings"
	"time"
)

type SortField string

const (
	SortByID        SortField = "id"
	SortByPriority  SortField = "priority"
	SortByDueDate   SortField = "due_date"
	SortByCreatedAt SortField = "created_at"
	SortByTitle     SortField = "title"
)

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type SortOrder struct {
	Field     SortField
	Direction SortDirection
}

type TaskFilter struct {
	Statuses   []Status
	Priorities []Priority
	Tags       []Tag
	ParentID   *string
	RootOnly   bool
	DueBefore  *time.Time
	DueAfter   *time.Time
	SearchTerm string
}

// FilterTasks applies the filter criteria in-memory to the task slice.
func FilterTasks(tasks []Task, filter TaskFilter) []Task {
	if len(tasks) == 0 {
		return []Task{}
	}

	if filter.RootOnly && filter.ParentID != nil {
		return []Task{}
	}

	searchTerm := strings.ToLower(strings.TrimSpace(filter.SearchTerm))

	result := make([]Task, 0, len(tasks))

	for _, task := range tasks {
		if !matchesFilter(task, filter, searchTerm) {
			continue
		}
		result = append(result, task)
	}

	return result
}

func matchesFilter(task Task, f TaskFilter, searchTerm string) bool {
	if len(f.Statuses) > 0 {
		matched := false
		for _, s := range f.Statuses {
			if task.Status == s {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.Priorities) > 0 {
		matched := false
		for _, p := range f.Priorities {
			if task.Priority == p {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.Tags) > 0 {
		for _, ft := range f.Tags {
			hasTag := false
			for _, tt := range task.Tags {
				if tt == ft {
					hasTag = true
					break
				}
			}
			if !hasTag {
				return false
			}
		}
	}

	if f.RootOnly && f.ParentID != nil {
		return false
	}
	if f.RootOnly {
		if task.ParentID != nil {
			return false
		}
	} else if f.ParentID != nil {
		if task.ParentID == nil || *task.ParentID != *f.ParentID {
			return false
		}
	}

	if f.DueBefore != nil {
		if task.DueDate == nil || !task.DueDate.Before(*f.DueBefore) {
			return false
		}
	}
	if f.DueAfter != nil {
		if task.DueDate == nil || !task.DueDate.After(*f.DueAfter) {
			return false
		}
	}

	if searchTerm != "" {
		if !strings.Contains(strings.ToLower(task.Title), searchTerm) {
			if task.Description == "" || !strings.Contains(strings.ToLower(task.Description), searchTerm) {
				return false
			}
		}
	}

	return true
}

// SortTasks sorts a slice of tasks in-place based on multi-key order criteria.
// An implicit deterministic final tie-breaker of SortByID ASC is always applied.
func SortTasks(tasks []Task, order []SortOrder) {
	if len(tasks) <= 1 {
		return
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		a := tasks[i]
		b := tasks[j]

		for _, ord := range order {
			cmp := compareTasksByField(a, b, ord.Field)
			if cmp == 0 {
				continue
			}
			if ord.Direction == SortDesc {
				return cmp > 0
			}
			return cmp < 0
		}

		return a.ID < b.ID
	})
}

func compareTasksByField(a, b Task, field SortField) int {
	switch field {
	case SortByID:
		if a.ID < b.ID {
			return -1
		} else if a.ID > b.ID {
			return 1
		}
		return 0

	case SortByPriority:
		wa := a.Priority.Weight()
		wb := b.Priority.Weight()
		if wa < wb {
			return -1
		} else if wa > wb {
			return 1
		}
		return 0

	case SortByDueDate:
		if a.DueDate == nil && b.DueDate == nil {
			return 0
		}
		if a.DueDate == nil {
			return 1
		}
		if b.DueDate == nil {
			return -1
		}
		if a.DueDate.Before(*b.DueDate) {
			return -1
		} else if a.DueDate.After(*b.DueDate) {
			return 1
		}
		return 0

	case SortByCreatedAt:
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		} else if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}
		return 0

	case SortByTitle:
		ta := strings.ToLower(a.Title)
		tb := strings.ToLower(b.Title)
		if ta < tb {
			return -1
		} else if ta > tb {
			return 1
		}
		return 0

	default:
		return 0
	}
}
