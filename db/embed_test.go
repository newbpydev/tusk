package db

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbed_ForwardOnly(t *testing.T) {
	names, err := fs.Glob(Migrations(), "migrations/*")
	if err != nil || len(names) != 1 || names[0] != "migrations/001_initial_schema.sql" {
		t.Fatalf("inventory=%v err=%v", names, err)
	}
	b, err := fs.ReadFile(Migrations(), names[0])
	if err != nil || !strings.Contains(string(b), "CREATE TABLE tasks") {
		t.Fatalf("schema unavailable: %v", err)
	}
}
