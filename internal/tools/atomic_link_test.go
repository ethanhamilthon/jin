package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRunSymlinkedParentWithRelativeLink(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	os.MkdirAll(filepath.Join(real, "sub"), 0o755)
	os.Symlink(filepath.Join(real, "sub"), filepath.Join(root, "alias"))
	os.Symlink("../target", filepath.Join(real, "sub", "link"))
	os.WriteFile(filepath.Join(real, "target"), []byte("old"), 0o644)
	os.WriteFile(filepath.Join(root, "target"), []byte("unrelated"), 0o644)
	args := `{"path":"` + filepath.Join(root, "alias", "link") + `","content":"new"}`
	if _, err := (Write{}).Run(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(real, "target")); string(data) != "new" {
		t.Fatalf("intended target=%q", data)
	}
	if data, _ := os.ReadFile(filepath.Join(root, "target")); string(data) != "unrelated" {
		t.Fatalf("unrelated file changed to %q", data)
	}
}

func TestWriteThroughDanglingLinkThenRevert(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "link.txt")
	os.Symlink("missing.txt", link)
	var changes []Change
	ctx := WithChangeSink(context.Background(), func(c Change) { changes = append(changes, c) })
	if _, err := (Write{}).Run(ctx, `{"path":"`+link+`","content":"hi"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := Revert(changes); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link must survive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "missing.txt")); !os.IsNotExist(err) {
		t.Fatalf("created target must be removed: %v", err)
	}
}
