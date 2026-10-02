package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONRequiresOneCompleteValue(t *testing.T) {
	for _, data := range []string{"{} {}", "{} null", "{} garbage", "{} ["} {
		t.Run(data, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "input.json")
			if err := os.WriteFile(file, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			var value manifest
			if err := readJSON(file, &value); err == nil {
				t.Fatal("trailing JSON/data accepted")
			}
		})
	}
	file := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(file, []byte("{} \n\t"), 0600); err != nil {
		t.Fatal(err)
	}
	var value manifest
	if err := readJSON(file, &value); err != nil {
		t.Fatalf("valid trailing whitespace: %v", err)
	}
}
