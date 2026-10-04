package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomeProjectHooksAreNotLoadedTwice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := Create("lint")
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(path, []byte("Run make lint."), 0o644)
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".jin"), 0o755)
	if err := os.Symlink(filepath.Dir(path), ProjectDir(dir)); err != nil {
		t.Fatal(err)
	}
	got := ActiveIn(dir, nil, true)
	if len(got) != 1 || got[0].Project {
		t.Fatalf("got %+v", got)
	}
	if names, _ := ListProject(dir); len(names) != 0 {
		t.Fatalf("project names = %v", names)
	}
}
