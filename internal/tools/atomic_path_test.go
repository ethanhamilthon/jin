package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndEditUseLiteralDotDotPath(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "real", "sub"), 0o755)
	os.Symlink(filepath.Join(root, "real", "sub"), filepath.Join(root, "alias"))
	writeFile(t, filepath.Join(root, "real", "target"), "old")
	writeFile(t, filepath.Join(root, "target"), "unrelated")
	literal := filepath.Join(root, "alias") + "/../target"
	if err := runWrite(literal, "written", context.Background()); err != nil {
		t.Fatal(err)
	}
	edit := `{"path":"` + literal + `","old_string":"written","new_string":"edited"}`
	if _, err := (Edit{}).Run(context.Background(), edit); err != nil {
		t.Fatal(err)
	}
	if readFile(t, filepath.Join(root, "real", "target")) != "edited" || readFile(t, filepath.Join(root, "target")) != "unrelated" {
		t.Fatal("OS path semantics were not kept")
	}
}

func TestTwoWritesUnderSymlinkedParentRevertTogether(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "real"), 0o755)
	os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "alias"))
	path := filepath.Join(root, "alias", "new", "file.txt")
	var changes []Change
	ctx := WithChangeSink(context.Background(), func(c Change) { changes = append(changes, c) })
	if err := runWrite(path, "one", ctx); err != nil {
		t.Fatal(err)
	}
	if err := runWrite(path, "two", ctx); err != nil {
		t.Fatal(err)
	}
	if changes[0].Path != changes[1].Path {
		t.Fatalf("unstable change paths: %q vs %q", changes[0].Path, changes[1].Path)
	}
	if _, err := Revert(changes); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "real", "new", "file.txt")); !os.IsNotExist(err) {
		t.Fatalf("created file must be removed: %v", err)
	}
}

func TestWriteThroughSymlinkThenRevert(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "real.txt"), "old")
	link := filepath.Join(dir, "link.txt")
	os.Symlink("real.txt", link)
	var changes []Change
	ctx := WithChangeSink(context.Background(), func(c Change) { changes = append(changes, c) })
	if err := runWrite(link, "new", ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := Revert(changes); err != nil {
		t.Fatal(err)
	}
	if !isLink(link) || readFile(t, filepath.Join(dir, "real.txt")) != "old" {
		t.Fatal("link must stay and target must be restored")
	}
}
