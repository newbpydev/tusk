package core_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
)

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		input   string
		want    core.Tag
		wantErr bool
	}{
		{"backend", core.Tag("backend"), false},
		{"#backend", core.Tag("backend"), false},
		{"  #Backend  ", core.Tag("backend"), false},
		{"frontend-v2", core.Tag("frontend-v2"), false},
		{"k8s", core.Tag("k8s"), false},
		{"ci-cd-pipeline", core.Tag("ci-cd-pipeline"), false},
		{"", "", true},
		{"#", "", true},
		{"   ", "", true},
		{"invalid tag", "", true},
		{"bad_tag", "", true},
		{"-leading-dash", "", true},
		{"trailing-dash-", "", true},
		{"double--dash", "", true},
		{"# backend", "", true},
		{"##backend", "", true},
		{"#--backend", "", true},
		{strings.Repeat("a", 33), "", true},
		{strings.Repeat("a", 32), core.Tag(strings.Repeat("a", 32)), false},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := core.NormalizeTag(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("NormalizeTag(%q) expected error, got nil", tc.input)
				}
				if !errors.Is(err, core.ErrInvalidTag) {
					t.Errorf("NormalizeTag(%q) error = %v, want %v", tc.input, err, core.ErrInvalidTag)
				}
			} else {
				if err != nil {
					t.Fatalf("NormalizeTag(%q) unexpected error: %v", tc.input, err)
				}
				if got != tc.want {
					t.Errorf("NormalizeTag(%q) = %q, want %q", tc.input, got, tc.want)
				}
				if got.String() != string(tc.want) {
					t.Errorf("Tag.String() = %q, want %q", got.String(), string(tc.want))
				}
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	raw := []string{"#frontend", " Backend ", "backend", "api-v1", "#api-v1"}
	want := []core.Tag{core.Tag("api-v1"), core.Tag("backend"), core.Tag("frontend")}

	got, err := core.NormalizeTags(raw)
	if err != nil {
		t.Fatalf("NormalizeTags unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeTags() = %v, want %v", got, want)
	}

	// With invalid tag
	_, err = core.NormalizeTags([]string{"valid", "invalid tag"})
	if err == nil || !errors.Is(err, core.ErrInvalidTag) {
		t.Errorf("NormalizeTags with invalid tag expected ErrInvalidTag, got %v", err)
	}

	// Empty slice
	empty, err := core.NormalizeTags(nil)
	if err != nil {
		t.Fatalf("NormalizeTags(nil) unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty slice, got %v", empty)
	}
}

func TestNormalizeTagSlice(t *testing.T) {
	raw := []core.Tag{core.Tag("frontend"), core.Tag("backend"), core.Tag("backend")}
	want := []core.Tag{core.Tag("backend"), core.Tag("frontend")}

	got, err := core.NormalizeTagSlice(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeTagSlice() = %v, want %v", got, want)
	}

	empty, err := core.NormalizeTagSlice(nil)
	if err != nil || len(empty) != 0 {
		t.Errorf("NormalizeTagSlice(nil) failed: %v", err)
	}
}
