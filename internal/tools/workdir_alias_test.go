package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDirSeenGuardFollowsSymlinkAlias(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.MkdirAll(filepath.Join(real, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(real, "sub"), filepath.Join(base, "alias")); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(real, "target")
	if err := os.WriteFile(target, []byte("before"), 0o644); err != nil {
		t.Fatal(err)
	}
	registry := BuildDir([]string{"read", "write"}, nil, base)
	read, _ := registry.Get("read")
	if _, err := read.Run(context.Background(), `{"path":"alias/../target"}`); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("changed outside tool"), 0o644); err != nil {
		t.Fatal(err)
	}
	write, _ := registry.Get("write")
	_, err := write.Run(context.Background(), `{"path":"alias/../target","content":"overwrite"}`)
	if err == nil || !strings.Contains(err.Error(), "changed on disk") {
		t.Fatalf("write through alias should be guarded, got %v", err)
	}
}

func TestBuildDirChangePathSupportsUndoThroughAlias(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.MkdirAll(filepath.Join(real, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(real, "sub"), filepath.Join(base, "alias")); err != nil {
		t.Fatal(err)
	}
	actual := filepath.Join(real, "sub", "new", "file.txt")
	var changes []Change
	ctx := WithChangeSink(context.Background(), func(c Change) { changes = append(changes, c) })
	registry := BuildDir([]string{"write"}, nil, base)
	write, _ := registry.Get("write")
	if _, err := write.Run(ctx, `{"path":"alias/new/file.txt","content":"created"}`); err != nil {
		t.Fatal(err)
	}
	wantPath := fileIdentity(actual)
	if len(changes) != 1 || changes[0].Path != wantPath || changes[0].Existed {
		t.Fatalf("change = %+v, want canonical new-file path %q", changes, wantPath)
	}
	if _, err := Revert(changes); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(actual); !os.IsNotExist(err) {
		t.Fatalf("undo left target in place: %v", err)
	}
	if info, err := os.Lstat(filepath.Join(base, "alias")); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("undo removed or replaced the alias: %v", err)
	}
}
