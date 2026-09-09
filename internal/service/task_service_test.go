package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
	"time"
)

type untouchedRepository struct{ ports.TaskRepository }

func TestServiceOptions_NoIO(t *testing.T) {
	for _, policy := range []bool{false, true} {
		o := Options{Clock: func() time.Time { panic("clock") }, NewID: func(time.Time) (string, error) { panic("id") }, Location: time.UTC, AutoCompleteParent: policy}
		s, err := NewTaskService(untouchedRepository{}, o)
		if err != nil || s == nil {
			t.Fatalf("%v %v", s, err)
		}
		for _, drop := range []string{"repo", "clock", "id", "location"} {
			bad := o
			var repo ports.TaskRepository = untouchedRepository{}
			switch drop {
			case "repo":
				repo = nil
			case "clock":
				bad.Clock = nil
			case "id":
				bad.NewID = nil
			case "location":
				bad.Location = nil
			}
			s, err := NewTaskService(repo, bad)
			if s != nil || !errors.Is(err, ports.ErrInvalidServiceOptions) {
				t.Fatalf("%s: %v", drop, err)
			}
		}
	}
}
func TestReferenceTimeAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := prepareCreate(ctx, ports.CreateTaskCommand{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := prepareUpdate(ctx, ports.UpdateTaskCommand{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := prepareQuery(ctx, ports.TaskQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, now := range []time.Time{{}, time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)} {
		if _, err := referenceTime(func() time.Time { return now }); !errors.Is(err, ports.ErrInvalidReferenceTime) {
			t.Fatal(err)
		}
	}
	calls := 0
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("x", 3600))
	got, err := referenceTime(func() time.Time { calls++; return now })
	if err != nil || calls != 1 || got.Location() != time.UTC || !got.Equal(now) {
		t.Fatalf("%v %v %d", got, err, calls)
	}
}
