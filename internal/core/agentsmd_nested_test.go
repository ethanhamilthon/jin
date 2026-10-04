package core

import (
	"path/filepath"
	"strings"
	"testing"
)

func nestedTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "AGENTS.md"), "root")
	writeFile(t, filepath.Join(root, "a", "AGENTS.md"), "rules a")
	writeFile(t, filepath.Join(root, "a", "b", "AGENTS.md"), "  ")
	writeFile(t, filepath.Join(root, "a", "b", "c", "AGENTS.md"), "rules c")
	writeFile(t, filepath.Join(root, "a", "b", "c", "f.go"), "x")
	writeFile(t, filepath.Join(root, "other", "f.go"), "x")
	return root
}

func TestNearestUnseenWalksDownOnce(t *testing.T) {
	root := nestedTree(t)
	sent := map[string]bool{}
	files := NearestUnseen(root, "a/b/c/f.go", sent)
	if len(files) != 2 || files[0].Content != "rules a" || files[1].Content != "rules c" {
		t.Fatalf("files = %+v", files)
	}
	if again := NearestUnseen(root, filepath.Join(root, "a", "b", "c", "f.go"), sent); len(again) != 0 {
		t.Fatalf("sent twice: %+v", again)
	}
}

func TestNearestUnseenIgnoresWorkdirAndOutside(t *testing.T) {
	root := nestedTree(t)
	for _, path := range []string{"f.go", "AGENTS.md", "other/f.go", filepath.Join(filepath.Dir(root), "x.go")} {
		if files := NearestUnseen(root, path, map[string]bool{}); len(files) != 0 {
			t.Errorf("%s: %+v", path, files)
		}
	}
}

func TestNestedAgentsAppendsOncePerAgentAndCuts(t *testing.T) {
	root := nestedTree(t)
	writeFile(t, filepath.Join(root, "big", "AGENTS.md"), strings.Repeat("z", 20000))
	t.Chdir(root)
	agent := NewAgent(nil, "sys", nil)
	call := readCall(t, "1", "a/b/c/f.go")
	first := agent.nestedAgents(call, "ok")
	if !strings.Contains(first, `<agents-md path="`+filepath.Join(root, "a", "AGENTS.md")+`" source="subdirectory">`) || !strings.Contains(first, "rules c") {
		t.Fatalf("first = %q", first)
	}
	if agent.nestedAgents(call, "ok") != "" || NewAgent(nil, "s", nil).nestedAgents(call, "ok") == "" {
		t.Fatal("set is not per agent")
	}
	if got := agent.nestedAgents(readCall(t, "2", "big/x"), "Error: nope"); got != "" {
		t.Fatalf("error result got %q", got)
	}
	big := agent.nestedAgents(readCall(t, "3", "big/x"), "ok")
	if len(big) > 8<<10+400 || !strings.Contains(big, "[truncated: 8192 of 20000 bytes shown]") {
		t.Fatalf("len %d, %q", len(big), big[len(big)-80:])
	}
}
