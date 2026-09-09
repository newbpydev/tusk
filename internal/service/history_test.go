package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
)

func TestHistory_EmptyMissingAndDetached(t *testing.T) {
	r := repository(t, task("a", "", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	ev, e := s.GetTaskHistory(context.Background(), "a")
	if e != nil || ev == nil || len(ev) != 0 {
		t.Fatalf("%v %v", ev, e)
	}
	if ev, e = s.GetTaskHistory(context.Background(), "missing"); !errors.Is(e, core.ErrTaskNotFound) || ev != nil {
		t.Fatalf("%v %v", ev, e)
	}
	if _, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Title: ptr("new")}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"}); e != nil {
		t.Fatal(e)
	}
	ev, e = s.GetTaskHistory(context.Background(), "a")
	if e != nil || len(ev) != 2 || ev[0].Sequence >= ev[1].Sequence {
		t.Fatalf("%v %v", ev, e)
	}
	ev[0].ChangedFields[0] = "changed"
	again, _ := s.GetTaskHistory(context.Background(), "a")
	if again[0].ChangedFields[0] != "title" {
		t.Fatal("aliased history")
	}
}
