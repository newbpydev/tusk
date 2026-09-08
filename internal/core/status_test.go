package core_test

import (
	"errors"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    core.Status
		wantErr bool
	}{
		{"todo", core.StatusTodo, false},
		{"in-progress", core.StatusInProgress, false},
		{"blocked", core.StatusBlocked, false},
		{"done", core.StatusDone, false},
		{" TODO ", core.StatusTodo, false},
		{"In-Progress", core.StatusInProgress, false},
		{"BLOCKED", core.StatusBlocked, false},
		{"Done", core.StatusDone, false},
		{"invalid", "", true},
		{"", "", true},
		{"review", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := core.ParseStatus(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseStatus(%q) expected error, got nil", tc.input)
				}
				if !errors.Is(err, core.ErrInvalidStatus) {
					t.Errorf("ParseStatus(%q) error = %v, want %v", tc.input, err, core.ErrInvalidStatus)
				}
			} else {
				if err != nil {
					t.Fatalf("ParseStatus(%q) unexpected error: %v", tc.input, err)
				}
				if got != tc.want {
					t.Errorf("ParseStatus(%q) = %v, want %v", tc.input, got, tc.want)
				}
				if !got.IsValid() {
					t.Errorf("expected %v to be valid", got)
				}
			}
		})
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	if !core.StatusDone.IsTerminal() {
		t.Errorf("expected StatusDone to be terminal")
	}
	if core.StatusTodo.IsTerminal() {
		t.Errorf("expected StatusTodo to not be terminal")
	}
	if core.StatusInProgress.IsTerminal() {
		t.Errorf("expected StatusInProgress to not be terminal")
	}
	if core.StatusBlocked.IsTerminal() {
		t.Errorf("expected StatusBlocked to not be terminal")
	}
}

func TestStatusTransitions(t *testing.T) {
	valid := []struct {
		from core.Status
		to   core.Status
	}{
		// From todo
		{core.StatusTodo, core.StatusTodo},
		{core.StatusTodo, core.StatusInProgress},
		{core.StatusTodo, core.StatusBlocked},
		{core.StatusTodo, core.StatusDone},

		// From in-progress
		{core.StatusInProgress, core.StatusInProgress},
		{core.StatusInProgress, core.StatusTodo},
		{core.StatusInProgress, core.StatusBlocked},
		{core.StatusInProgress, core.StatusDone},

		// From blocked
		{core.StatusBlocked, core.StatusBlocked},
		{core.StatusBlocked, core.StatusTodo},
		{core.StatusBlocked, core.StatusInProgress},
		{core.StatusBlocked, core.StatusDone},

		// From done (reopening)
		{core.StatusDone, core.StatusDone},
		{core.StatusDone, core.StatusTodo},
		{core.StatusDone, core.StatusInProgress},
	}

	for _, tc := range valid {
		t.Run(string(tc.from)+"->"+string(tc.to), func(t *testing.T) {
			if !tc.from.CanTransitionTo(tc.to) {
				t.Errorf("expected %v -> %v to be valid", tc.from, tc.to)
			}
		})
	}

	invalid := []struct {
		from core.Status
		to   core.Status
	}{
		{core.StatusDone, core.StatusBlocked},
		{core.Status("unknown"), core.StatusTodo},
		{core.StatusTodo, core.Status("unknown")},
	}

	for _, tc := range invalid {
		t.Run("invalid_"+string(tc.from)+"->"+string(tc.to), func(t *testing.T) {
			if tc.from.CanTransitionTo(tc.to) {
				t.Errorf("expected %v -> %v to be invalid", tc.from, tc.to)
			}
		})
	}
}
