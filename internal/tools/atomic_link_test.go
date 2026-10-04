package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	return string(data)
}

func isLink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func runWrite(path, content string, ctx context.Context) error {
	_, err := (Write{}).Run(ctx, `{"path":"`+path+`","content":"`+content+`"}`)
	return err
}

func TestWriteKeepsSymlinkAndChangesTarget(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "real.txt"), "old")
	link := filepath.Join(dir, "link.txt")
	os.Symlink("real.txt", link)
	if err := runWrite(link, "new", context.Background()); err != nil {
		t.Fatal(err)
	}
	if !isLink(link) || readFile(t, filepath.Join(dir, "real.txt")) != "new" {
		t.Fatal("link must stay and target must change")
	}
}

func TestWriteFollowsLinkChain(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "c.txt"), "old")
	os.Symlink("b", filepath.Join(dir, "a"))
	os.Symlink("c.txt", filepath.Join(dir, "b"))
	if err := runWrite(filepath.Join(dir, "a"), "x", context.Background()); err != nil {
		t.Fatal(err)
	}
	if readFile(t, filepath.Join(dir, "c.txt")) != "x" || !isLink(filepath.Join(dir, "a")) {
		t.Fatal("chain end must change, links stay")
	}
}

func TestWriteRefusesBrokenAndLoopedSymlinks(t *testing.T) {
	dir := t.TempDir()
	os.Symlink("missing.txt", filepath.Join(dir, "dangling"))
	os.Symlink("l2", filepath.Join(dir, "l1"))
	os.Symlink("l1", filepath.Join(dir, "l2"))
	for _, name := range []string{"dangling", "l1"} {
		err := runWrite(filepath.Join(dir, name), "x", context.Background())
		if err == nil || !strings.Contains(err.Error(), "refusing to write through a broken symlink") {
			t.Fatalf("%s: err=%v", name, err)
		}
		if !isLink(filepath.Join(dir, name)) {
			t.Fatalf("%s: link was replaced", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "missing.txt")); !os.IsNotExist(err) {
		t.Fatal("nothing must be created")
	}
}

func TestWriteSymlinkedParentWithRelativeLink(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "real", "sub"), 0o755)
	os.Symlink(filepath.Join(root, "real", "sub"), filepath.Join(root, "alias"))
	os.Symlink("../target", filepath.Join(root, "real", "sub", "link"))
	writeFile(t, filepath.Join(root, "real", "target"), "old")
	writeFile(t, filepath.Join(root, "target"), "unrelated")
	if err := runWrite(filepath.Join(root, "alias", "link"), "new", context.Background()); err != nil {
		t.Fatal(err)
	}
	if readFile(t, filepath.Join(root, "real", "target")) != "new" || readFile(t, filepath.Join(root, "target")) != "unrelated" {
		t.Fatal("wrong file written")
	}
}
