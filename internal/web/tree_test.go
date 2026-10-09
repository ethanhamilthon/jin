package web

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, text := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestListTreeSortsAndHidesGitAndIgnored(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"b.go": "x", "A.md": "x", "src/main.go": "x", "dist/out.js": "x", ".gitignore": "dist/\n*.log\n", "run.log": "x"})
	if err := exec.Command("git", "-C", root, "init", "-q").Run(); err != nil {
		t.Skip("git is not available")
	}
	list, err := listTree(context.Background(), root, "", false)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range list {
		names = append(names, e.Name)
	}
	if got := strings.Join(names, ","); got != "src,.gitignore,A.md,b.go" {
		t.Fatalf("default list = %s", got)
	}
	all, _ := listTree(context.Background(), root, "", true)
	if len(all) != 7 {
		t.Fatalf("with hidden: %d entries", len(all))
	}
}

func TestProjectPathsCannotLeaveTheProject(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeFiles(t, root, map[string]string{"ok.txt": "hi"})
	writeFiles(t, outside, map[string]string{"secret.txt": "no"})
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks are not available")
	}
	for _, rel := range []string{"../" + filepath.Base(outside), "link/secret.txt", "link", "/etc/passwd"} {
		if _, err := inProject(root, rel); err == nil {
			t.Errorf("%q was allowed", rel)
		}
	}
	if _, err := inProject(root, "ok.txt"); err != nil {
		t.Errorf("a file of the project was refused: %v", err)
	}
}

func TestViewFileKinds(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"a.go": "package a\n", "notes.md": "# hi", "bin.dat": "ab\x00cd"})
	if err := os.WriteFile(filepath.Join(root, "big.txt"), make([]byte, maxPreview+1), 0o644); err != nil {
		t.Fatal(err)
	}
	png := "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"
	writeFiles(t, root, map[string]string{"p.png": png})
	want := map[string]string{"a.go": "text", "notes.md": "markdown", "bin.dat": "binary", "big.txt": "large", "p.png": "image"}
	for name, kind := range want {
		view, err := viewFile(root, name)
		if err != nil || view.Kind != kind {
			t.Errorf("%s: kind %q, err %v, want %q", name, view.Kind, err, kind)
		}
	}
	if view, _ := viewFile(root, "a.go"); view.Content != "package a\n" {
		t.Errorf("content = %q", view.Content)
	}
	if view, _ := viewFile(root, "big.txt"); view.Content != "" {
		t.Error("a file over 1 MB was read")
	}
}
