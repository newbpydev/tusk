package core_test

import (
	"errors"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
)

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input      string
		want       core.Priority
		wantWeight int
		wantString string
		wantErr    bool
	}{
		{"low", core.PriorityLow, 1, "low", false},
		{"1", core.PriorityLow, 1, "low", false},
		{"medium", core.PriorityMedium, 2, "medium", false},
		{"2", core.PriorityMedium, 2, "medium", false},
		{"high", core.PriorityHigh, 3, "high", false},
		{"3", core.PriorityHigh, 3, "high", false},
		{"urgent", core.PriorityUrgent, 4, "urgent", false},
		{"4", core.PriorityUrgent, 4, "urgent", false},
		{" LOW ", core.PriorityLow, 1, "low", false},
		{"Urgent", core.PriorityUrgent, 4, "urgent", false},
		{"0", 0, 0, "", true},
		{"5", 0, 0, "", true},
		{"critical", 0, 0, "", true},
		{"", 0, 0, "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := core.ParsePriority(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParsePriority(%q) expected error, got nil", tc.input)
				}
				if !errors.Is(err, core.ErrInvalidPriority) {
					t.Errorf("ParsePriority(%q) error = %v, want %v", tc.input, err, core.ErrInvalidPriority)
				}
			} else {
				if err != nil {
					t.Fatalf("ParsePriority(%q) unexpected error: %v", tc.input, err)
				}
				if got != tc.want {
					t.Errorf("ParsePriority(%q) = %v, want %v", tc.input, got, tc.want)
				}
				if got.Weight() != tc.wantWeight {
					t.Errorf("Priority(%v).Weight() = %d, want %d", got, got.Weight(), tc.wantWeight)
				}
				if got.String() != tc.wantString {
					t.Errorf("Priority(%v).String() = %q, want %q", got, got.String(), tc.wantString)
				}
				if !got.IsValid() {
					t.Errorf("expected priority %v to be valid", got)
				}
			}
		})
	}
}

func TestPriority_IsValid(t *testing.T) {
	if core.Priority(0).IsValid() {
		t.Errorf("expected Priority(0) to be invalid")
	}
	if core.Priority(5).IsValid() {
		t.Errorf("expected Priority(5) to be invalid")
	}
	if !core.PriorityLow.IsValid() {
		t.Errorf("expected PriorityLow to be valid")
	}
	if !core.PriorityUrgent.IsValid() {
		t.Errorf("expected PriorityUrgent to be valid")
	}
}
