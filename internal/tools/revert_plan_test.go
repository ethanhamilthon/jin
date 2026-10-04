package tools

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPlanRevertSplitsRestorableFromChanged(t *testing.T) {
	dir := t.TempDir()
	same := filepath.Join(dir, "same.txt")
	moved := filepath.Join(dir, "moved.txt")
	gone := filepath.Join(dir, "gone.txt")
	_ = os.WriteFile(same, []byte("new"), 0o644)
	_ = os.WriteFile(moved, []byte("edited later"), 0o644)
	changes := []Change{
		{Path: same, Existed: true, Before: "old", After: "new"},
		{Path: moved, Existed: true, Before: "old", After: "new"},
		{Path: gone, Existed: true, Before: "old", After: "new"},
	}
	plan, err := PlanRevert(changes)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{same}; !reflect.DeepEqual(plan.Restored, want) {
		t.Errorf("restored = %v", plan.Restored)
	}
	if want := []string{moved, gone}; !reflect.DeepEqual(plan.Skipped, want) {
		t.Errorf("skipped = %v", plan.Skipped)
	}
	if data, _ := os.ReadFile(same); string(data) != "new" {
		t.Errorf("plan wrote the file: %q", data)
	}
}
