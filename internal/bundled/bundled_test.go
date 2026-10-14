package bundled

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestPathMacLayout(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "Jin.app", "Contents", "MacOS", "Jin")
	want := filepath.Join(root, "Jin.app", "Contents", "Resources", "jin")
	touch(t, exe)
	if _, ok := Path(exe); ok {
		t.Fatal("found a CLI that does not exist")
	}
	touch(t, want)
	if got, ok := Path(exe); !ok || got != want {
		t.Fatalf("got %q, %v", got, ok)
	}
}

func TestPathLinuxLayout(t *testing.T) {
	dir := t.TempDir()
	exe, want := filepath.Join(dir, "jin-desktop"), filepath.Join(dir, "jin")
	touch(t, exe)
	touch(t, want)
	if got, ok := Path(exe); !ok || got != want {
		t.Fatalf("got %q, %v", got, ok)
	}
}

func TestPathIgnoresOtherBinaries(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "jin"))
	touch(t, filepath.Join(dir, "other"))
	if _, ok := Path(filepath.Join(dir, "other")); ok {
		t.Fatal("a plain binary has no bundled CLI")
	}
}

func TestInside(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "Jin.app", "Contents", "Resources", "jin")
	linux := filepath.Join(dir, "pkg", "jin")
	alone := filepath.Join(dir, "bin", "jin")
	touch(t, linux)
	touch(t, filepath.Join(dir, "pkg", "jin-desktop"))
	touch(t, alone)
	for path, want := range map[string]bool{app: true, linux: true, alone: false} {
		if got := Inside(path); got != want {
			t.Errorf("Inside(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestAddToPathPutsTheBundledFolderFirst(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "jin-desktop")
	touch(t, exe)
	touch(t, filepath.Join(dir, "jin"))
	t.Setenv("PATH", "/usr/bin")
	AddToPath(exe)
	if got, want := os.Getenv("PATH"), dir+string(os.PathListSeparator)+"/usr/bin"; got != want {
		t.Fatalf("PATH = %q, want %q", got, want)
	}
}

func TestAddToPathWithoutBundledCLIKeepsPath(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "jin-desktop")
	touch(t, exe)
	t.Setenv("PATH", "/usr/bin")
	AddToPath(exe)
	if got := os.Getenv("PATH"); got != "/usr/bin" {
		t.Fatalf("PATH = %q", got)
	}
}
