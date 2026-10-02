package prompts

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestEnsureDefaultsCreatesAllThree(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := EnsureDefaults(); err != nil {
		t.Fatal(err)
	}
	names, _ := List()
	if !slices.Equal(names, []string{"plan", "review", "subagents"}) {
		t.Fatalf("names = %v", names)
	}
	path, _ := Path("plan")
	if body, _ := os.ReadFile(path); !strings.HasPrefix(string(body), "PLAN MODE") {
		t.Fatalf("plan body = %q", body)
	}
}

func TestEnsureDefaultsNeverTouchesExistingFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := EnsureDefaults(); err != nil {
		t.Fatal(err)
	}
	plan, _ := Path("plan")
	review, _ := Path("review")
	if err := os.WriteFile(plan, []byte("my plan"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(review, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDefaults(); err != nil {
		t.Fatal(err)
	}
	if body, _ := os.ReadFile(plan); string(body) != "my plan" {
		t.Errorf("edited file changed: %q", body)
	}
	if body, _ := os.ReadFile(review); len(body) != 0 {
		t.Errorf("empty file refilled: %q", body)
	}
}

func TestEnsureDefaultsRestoresDeletedFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_ = EnsureDefaults()
	path, _ := Path("subagents")
	_ = os.Remove(path)
	if err := EnsureDefaults(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("not restored: %v", err)
	}
}

// The prompts name real flags; fail when one is renamed.
func TestSubagentsPromptUsesRealCommands(t *testing.T) {
	body, err := defaults.ReadFile("defaults/subagents.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"jin models", "jin refresh-models", "--no-session", "--model", "--timeout", "--tools", "--effort", "ask_user", "todo", ".exit", ".pid", "sleep", "/dev/null", "no speedup", "smart:", "fast:"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("subagents prompt lacks %q", want)
		}
	}
}
