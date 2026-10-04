package tools

import (
	"context"
	"os"
	"path/filepath"
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

func TestWriteFileAtomicKeepsSymlink(t *testing.T) {
	dir := t.TempDir()
	real, link := filepath.Join(dir, "real.txt"), filepath.Join(dir, "link.txt")
	os.WriteFile(real, []byte("old"), 0o644)
	os.Symlink("real.txt", link)
	if err := writeFileAtomic(link, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Lstat(link); info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink was replaced")
	}
	if data, _ := os.ReadFile(real); string(data) != "new" {
		t.Fatalf("target=%q", data)
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

func TestWriteFileAtomicFillsDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "link.txt")
	os.Symlink("missing.txt", link)
	if err := writeFileAtomic(link, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Lstat(link); info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink was replaced")
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "missing.txt")); string(data) != "new" {
		t.Fatalf("target=%q", data)
	}
}

func TestWriteFileAtomicFollowsLinkChain(t *testing.T) {
	dir := t.TempDir()
	os.Symlink("b", filepath.Join(dir, "a"))
	os.Symlink("c.txt", filepath.Join(dir, "b"))
	if err := writeFileAtomic(filepath.Join(dir, "a"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "c.txt")); string(data) != "x" {
		t.Fatalf("target=%q", data)
	}
}

func TestWriteFileAtomicRejectsSymlinkLoop(t *testing.T) {
	dir := t.TempDir()
	os.Symlink("b", filepath.Join(dir, "a"))
	os.Symlink("a", filepath.Join(dir, "b"))
	if err := writeFileAtomic(filepath.Join(dir, "a"), []byte("x"), 0o644); err == nil {
		t.Fatal("expected loop error")
	}
}

func TestWriteRunFillsDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "link.txt")
	os.Symlink("missing.txt", link)
	args := `{"path":"` + link + `","content":"hi"}`
	if _, err := (Write{}).Run(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Lstat(link); info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink was replaced")
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "missing.txt")); string(data) != "hi" {
		t.Fatalf("target=%q", data)
	}
}
