package prompts

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSystemPrompts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, name := range []string{"plan", "review", "subagents"} {
		if !IsSystem(name) {
			t.Errorf("IsSystem(%q) = false", name)
		}
		if _, err := Create(name); err == nil {
			t.Errorf("Create(%q) should fail", name)
		}
		if _, err := Path(name); err == nil {
			t.Errorf("Path(%q) should fail", name)
		}
		if err := Delete(name); err == nil {
			t.Errorf("Delete(%q) should fail", name)
		}
	}
	dir, _ := root()
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "plan.md"), []byte("user plan"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "my-prompt.md"), []byte("user prompt"), 0o644)

	infos, err := ListInfo()
	if err != nil {
		t.Fatal(err)
	}
	names, err := List()
	if err != nil || !slices.Equal(names, []string{"plan", "review", "subagents", "my-prompt"}) {
		t.Fatalf("List = %v, %v", names, err)
	}
	if len(infos) != 4 || !infos[0].System || infos[3].System {
		t.Fatalf("ListInfo = %+v", infos)
	}
	expanded := Expand("run #plan", Bodies(nil))
	if strings.Contains(expanded, "user plan") || !strings.Contains(expanded, "PLAN MODE") {
		t.Fatalf("Expand should use embedded system plan, got:\n%s", expanded)
	}
}

func TestBodiesSkipDisabledAndEmptyPrompts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, _ := root()
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "mine.md"), []byte("  my text \n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "blank.md"), []byte("  \n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "off.md"), []byte("hidden"), 0o644)
	bodies := Bodies([]string{"off", "review"})
	if bodies["mine"] != "my text" {
		t.Errorf("mine = %q", bodies["mine"])
	}
	for _, gone := range []string{"blank", "off", "review"} {
		if _, ok := bodies[gone]; ok {
			t.Errorf("%q must not be in the bodies", gone)
		}
	}
	if bodies["plan"] == "" || bodies["subagents"] == "" {
		t.Error("enabled system prompts must be there")
	}
}

func TestExpandUsesTheGivenBodiesOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got := Expand("use #style and #plan", map[string]string{"style": "Filled in text."})
	if !strings.Contains(got, "<prompt name=\"style\">\nFilled in text.\n</prompt>") || strings.Contains(got, "name=\"plan\"") {
		t.Errorf("Expand = %q", got)
	}
}
