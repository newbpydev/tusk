package service

import (
	"github.com/newbpydev/tusk/internal/ports"
	"time"
)

// Options dependencies must support concurrent callers. The caller owns storage
// lifetime; construction does not call any dependency or perform I/O.
type Options struct {
	Clock              func() time.Time
	NewID              func(time.Time) (string, error)
	Location           *time.Location
	AutoCompleteParent bool
}
type TaskService struct {
	repo    ports.TaskRepository
	options Options
}

func NewTaskService(repo ports.TaskRepository, options Options) (*TaskService, error) {
	if repo == nil || options.Clock == nil || options.NewID == nil || options.Location == nil {
		return nil, ports.ErrInvalidServiceOptions
	}
	return &TaskService{repo: repo, options: options}, nil
}
func referenceTime(clock func() time.Time) (time.Time, error) {
	t := clock().Round(0).UTC()
	if !validTime(t) || t.IsZero() {
		return time.Time{}, ports.ErrInvalidReferenceTime
	}
	return t, nil
}
func validTime(t time.Time) bool { return t.UTC().Year() >= 1 && t.UTC().Year() <= 9999 }
