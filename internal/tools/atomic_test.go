package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomicKeepsModeAndReplaces(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.sh")
	os.WriteFile(path, []byte("old"), 0o755)
	os.Chmod(path, 0o755)
	if err := writeFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	data, _ := os.ReadFile(path)
	if string(data) != "new" || info.Mode().Perm() != 0o755 {
		t.Fatalf("data=%q mode=%v", data, info.Mode())
	}
}

func TestWriteFileAtomicNewFileDefaultMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "n.txt")
	if err := writeFileAtomic(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o644 {
		t.Fatalf("mode=%v", info.Mode())
	}
}

func TestWriteFileAtomicFailureLeavesOldFileAndNoTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	os.WriteFile(path, []byte("old"), 0o644)
	os.Chmod(dir, 0o500)
	defer os.Chmod(dir, 0o755)
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	if err := writeFileAtomic(path, []byte("new"), 0o644); err == nil {
		t.Fatal("expected error")
	}
	os.Chmod(dir, 0o755)
	if data, _ := os.ReadFile(path); string(data) != "old" {
		t.Fatalf("file=%q", data)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("leftover files: %v", entries)
	}
}

func TestWriteFileAtomicRenameFailureCleansTemp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "d")
	os.Mkdir(target, 0o755)
	os.WriteFile(filepath.Join(target, "x"), nil, 0o644)
	if err := writeFileAtomic(target, []byte("new"), 0o644); err == nil {
		t.Fatal("expected error replacing a directory")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("leftover files: %v", entries)
	}
}

func TestWriteFileAtomicLongBasename(t *testing.T) {
	path := filepath.Join(t.TempDir(), strings.Repeat("n", 250)+".txt")
	if err := writeFileAtomic(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); string(data) != "x" {
		t.Fatalf("data=%q", data)
	}
}
