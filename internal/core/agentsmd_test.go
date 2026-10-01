package core

import (
	"path/filepath"
	"testing"
)

func TestContextFilesOrderAndSkipping(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	mid := filepath.Join(root, "mid")
	dir := filepath.Join(mid, "project")
	global := filepath.Join(home, ".jin-dev", "AGENTS.md")
	writeFile(t, global, "global")
	writeFile(t, filepath.Join(root, "AGENTS.md"), "outer")
	writeFile(t, filepath.Join(mid, "AGENTS.md"), "  \n")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "project")
	files := ContextFiles(dir)
	want := []struct {
		kind    ContextKind
		content string
	}{{ContextGlobal, "global"}, {ContextParent, "outer"}, {ContextProject, "project"}}
	if len(files) != len(want) {
		t.Fatalf("got %+v", files)
	}
	for i, w := range want {
		if files[i].Kind != w.kind || files[i].Content != w.content {
			t.Errorf("file %d = %+v, want %+v", i, files[i], w)
		}
	}
}

func TestContextFilesDedupe(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(home, ".jin-dev", "AGENTS.md"), "global")
	if files := ContextFiles(filepath.Join(home, ".jin-dev")); len(files) != 1 {
		t.Fatalf("same file listed twice: %+v", files)
	}
}

func TestAgentsFilesKeepEmptyOnesAndCreateOnlyOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(t.TempDir(), "project")
	writeFile(t, filepath.Join(dir, "x"), "")
	writeFile(t, filepath.Join(home, ".jin-dev", "AGENTS.md"), "")
	files := AgentsFiles(dir)
	if len(files) != 1 || files[0].Kind != ContextGlobal {
		t.Fatalf("empty global file should be listed: %+v", files)
	}
	path, err := CreateProjectAgents(dir)
	if err != nil || path != filepath.Join(dir, "AGENTS.md") {
		t.Fatalf("create: %q, %v", path, err)
	}
	if files = AgentsFiles(dir); len(files) != 2 || files[1].Kind != ContextProject {
		t.Errorf("created file should be listed: %+v", files)
	}
	if _, err := CreateProjectAgents(dir); err == nil {
		t.Error("second create should fail")
	}
	if len(ContextFiles(dir)) != 0 {
		t.Error("empty files must stay out of the system prompt")
	}
}
