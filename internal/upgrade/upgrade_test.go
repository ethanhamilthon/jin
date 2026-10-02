package upgrade

import (
	"os"
	"path/filepath"
	"testing"

	"jin/internal/paths"
	"jin/internal/store"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestOldDefaultPromptsMoveToBackupOnce(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	prompts, _ := paths.Global("prompts")
	backup, _ := paths.Global("prompts.bak")
	write(t, filepath.Join(prompts, "plan.md"), "my edited plan")
	write(t, filepath.Join(prompts, "subagents.md"), "smart: gpt-6-sol")
	write(t, filepath.Join(prompts, "mine.md"), "keep me")

	if err := Run(db); err != nil {
		t.Fatal(err)
	}
	for _, gone := range []string{"plan.md", "subagents.md"} {
		if _, err := os.Stat(filepath.Join(prompts, gone)); !os.IsNotExist(err) {
			t.Errorf("%s must leave the prompts folder", gone)
		}
	}
	if data, _ := os.ReadFile(filepath.Join(backup, "plan.md")); string(data) != "my edited plan" {
		t.Errorf("backup of plan.md = %q", data)
	}
	if data, _ := os.ReadFile(filepath.Join(backup, "subagents.md")); string(data) != "smart: gpt-6-sol" {
		t.Errorf("backup of subagents.md = %q", data)
	}
	if data, _ := os.ReadFile(filepath.Join(prompts, "mine.md")); string(data) != "keep me" {
		t.Error("a user prompt must stay")
	}

	// A second run does nothing, even when a file with an old name appears.
	write(t, filepath.Join(prompts, "plan.md"), "new file of the user")
	if err := Run(db); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(prompts, "plan.md")); string(data) != "new file of the user" {
		t.Error("the migration must run only once")
	}
}

func TestBackupDoesNotOverwriteAnEarlierOne(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	prompts, _ := paths.Global("prompts")
	backup, _ := paths.Global("prompts.bak")
	write(t, filepath.Join(backup, "plan.md"), "older backup")
	write(t, filepath.Join(prompts, "plan.md"), "newer")
	if err := Run(db); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(backup, "plan.md")); string(data) != "older backup" {
		t.Errorf("older backup was overwritten: %q", data)
	}
	if data, _ := os.ReadFile(filepath.Join(backup, "plan.1.md")); string(data) != "newer" {
		t.Errorf("plan.1.md = %q", data)
	}
}

func TestFreshInstallHasNothingToMove(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Run(db); err != nil {
		t.Fatal(err)
	}
	if backup, _ := paths.Global("prompts.bak"); func() bool { _, err := os.Stat(backup); return err == nil }() {
		t.Error("no backup folder is needed when there was nothing to move")
	}
}
