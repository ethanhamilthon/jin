package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func toolArgs(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestBuildDirIsolatesConcurrentSessions(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	registries := []*Registry{
		BuildDir([]string{"read", "write", "edit"}, first),
		BuildDir([]string{"read", "write", "edit"}, second),
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	inputs := []string{
		toolArgs(t, map[string]string{"path": "note.txt", "content": "a"}),
		toolArgs(t, map[string]string{"path": "note.txt", "content": "b"}),
	}
	for i, registry := range registries {
		wg.Add(1)
		go func(i int, registry *Registry) {
			defer wg.Done()
			tool, _ := registry.Get("write")
			_, err := tool.Run(context.Background(), inputs[i])
			errs <- err
		}(i, registry)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, registry := range registries {
		tool, _ := registry.Get("read")
		got, err := tool.Run(context.Background(), `{"path":"note.txt"}`)
		if err != nil || !strings.Contains(got, string(rune('a'+i))) {
			t.Fatalf("session %d read %q, %v", i, got, err)
		}
		edit, _ := registry.Get("edit")
		args := toolArgs(t, map[string]string{"path": "note.txt", "old_string": string(rune('a' + i)), "new_string": "edited"})
		if summary, ok := edit.Summary(args); !ok || summary != "note.txt:1" {
			t.Fatalf("session %d edit summary = %q, %v", i, summary, ok)
		}
		if _, err := edit.Run(context.Background(), args); err != nil {
			t.Fatalf("session %d edit: %v", i, err)
		}
	}
	if _, err := os.Stat(filepath.Join(first, "note.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(second, "note.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestBuildDirHonorsAbsolutePathsAndBashPwd(t *testing.T) {
	workdir, outside := t.TempDir(), filepath.Join(t.TempDir(), "absolute.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	registry := BuildDir([]string{"read", "bash"}, workdir)
	read, _ := registry.Get("read")
	if got, err := read.Run(context.Background(), toolArgs(t, map[string]string{"path": outside})); err != nil || !strings.Contains(got, "outside") {
		t.Fatalf("absolute read = %q, %v", got, err)
	}
	bash, _ := registry.Get("bash")
	got, err := bash.Run(context.Background(), `{"command":"pwd"}`)
	wantDir, _ := filepath.EvalSymlinks(workdir)
	if err != nil || strings.TrimSpace(got) != wantDir {
		t.Fatalf("bash pwd = %q, %v; want %q", got, err, wantDir)
	}
}
