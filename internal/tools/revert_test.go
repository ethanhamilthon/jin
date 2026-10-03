package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRevertRestoresAndSkipsChangedFiles(t *testing.T) {
	dir := t.TempDir()
	edited := filepath.Join(dir, "a.txt")
	created := filepath.Join(dir, "new.txt")
	touched := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(edited, []byte("v3"), 0o644)
	_ = os.WriteFile(created, []byte("fresh"), 0o644)
	_ = os.WriteFile(touched, []byte("user edit"), 0o644)
	changes := []Change{
		{Path: edited, Existed: true, Before: "v1", After: "v2"},
		{Path: created, Before: "", After: "fresh"},
		{Path: edited, Existed: true, Before: "v2", After: "v3"},
		{Path: touched, Existed: true, Before: "old", After: "agent"},
	}
	result, err := Revert(changes)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Restored) != 2 || len(result.Skipped) != 1 || result.Skipped[0] != touched {
		t.Fatalf("result = %+v", result)
	}
	if data, _ := os.ReadFile(edited); string(data) != "v1" {
		t.Errorf("edited = %q", data)
	}
	if _, err := os.Stat(created); !os.IsNotExist(err) {
		t.Error("created file must be removed")
	}
	if data, _ := os.ReadFile(touched); string(data) != "user edit" {
		t.Errorf("touched = %q", data)
	}
}
