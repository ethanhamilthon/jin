package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileIdentityResolvesMissingSuffixThroughAliases(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	throughAlias := filepath.Join(alias, "missing", "nested", "file.txt")
	direct := filepath.Join(real, "missing", "nested", "file.txt")
	if got, want := fileIdentity(throughAlias), fileIdentity(direct); got != want {
		t.Fatalf("missing-path identities differ: %q != %q", got, want)
	}
}

func TestBuildDirSeenRejectsRetargetedSymlink(t *testing.T) {
	base, first, second := t.TempDir(), t.TempDir(), t.TempDir()
	path := filepath.Join(base, "alias")
	if err := os.Symlink(first, path); err != nil {
		t.Fatal(err)
	}
	for dir, content := range map[string]string{first: "first", second: "second"} {
		if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	registry := BuildDir([]string{"read", "write"}, nil, base)
	read, _ := registry.Get("read")
	if _, err := read.Run(context.Background(), `{"path":"alias/file.txt"}`); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(second, path); err != nil {
		t.Fatal(err)
	}
	write, _ := registry.Get("write")
	if _, err := write.Run(context.Background(), `{"path":"alias/file.txt","content":"unsafe"}`); err == nil {
		t.Fatal("write must reject an alias that now resolves to another file")
	}
	if got, err := os.ReadFile(filepath.Join(second, "file.txt")); err != nil || string(got) != "second" {
		t.Fatalf("retargeted file changed: %q, %v", got, err)
	}
}
