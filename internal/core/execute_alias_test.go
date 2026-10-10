package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"jin/internal/provider"
)

func TestGroupingResolvesSessionAliases(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(dir, "alias")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(real, "file"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	call := func(tool, path string) provider.ToolCall {
		args, _ := json.Marshal(map[string]string{"path": path})
		return fakeCall(path, tool, string(args))
	}
	for _, path := range []string{filepath.Join(real, "file"), "alias/file"} {
		calls := []provider.ToolCall{call("write", "real/file"), call("edit", path)}
		if got := len(nextGroup(calls, dir)); got != 1 {
			t.Fatalf("alias %s grouped %d calls", path, got)
		}
	}
	calls := []provider.ToolCall{call("write", "real/new"), call("write", "alias/new")}
	if len(nextGroup(calls, dir)) != 1 {
		t.Fatal("new file aliases must serialize")
	}
}

func TestGroupingPreservesSymlinkParentMeaning(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "real", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "real", "sub"), filepath.Join(dir, "alias")); err != nil {
		t.Fatal(err)
	}
	calls := []provider.ToolCall{
		fakeCall("1", "write", `{"path":"alias/../file"}`),
		fakeCall("2", "write", `{"path":"real/file"}`),
	}
	if len(nextGroup(calls, dir)) != 1 {
		t.Fatal("symlink/.. must resolve through the filesystem")
	}
}
