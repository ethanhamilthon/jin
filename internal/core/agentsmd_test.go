package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContextFilesTakesOnlyTheProjectFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	dir := filepath.Join(root, "project")
	writeFile(t, filepath.Join(home, ".jin-dev", "AGENTS.md"), "global")
	writeFile(t, filepath.Join(root, "AGENTS.md"), "outer")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "project")
	files := ContextFiles(dir)
	if len(files) != 1 || files[0].Kind != ContextProject || files[0].Content != "project" {
		t.Fatalf("got %+v", files)
	}
	writeFile(t, filepath.Join(dir, "AGENTS.md"), " \n")
	if files := ContextFiles(dir); len(files) != 0 {
		t.Fatalf("an empty file is listed: %+v", files)
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
