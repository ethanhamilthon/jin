package ui

import (
	"os"
	"strings"
	"testing"

	"jin/internal/sysprompt"
)

func TestSystemPromptCommandExistsAndCreatesTheFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cmd, ok := slashByName("system-prompt")
	if !ok || cmd.run == nil || cmd.desc == "" {
		t.Fatalf("/system-prompt is missing or incomplete: %+v", cmd)
	}
	a, _ := layoutApp(t)
	a.cfg.Editor = "" // no editor chosen: the command asks for one, after the file exists
	cmd.run(a, "")
	path, _ := sysprompt.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the file was not created: %v", err)
	}
	for _, want := range []string{"# system\n", "# compact\n", "# handoff\n"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("file lacks %q", want)
		}
	}
	if a.sel == nil {
		t.Error("without an editor the command must ask which one to use")
	}
}

func TestAsyncTasksCommandIsAlwaysThere(t *testing.T) {
	a, _ := layoutApp(t)
	a.cfg.HooksDisabled = []string{"async", "docs"}
	a.active.input, a.active.cursor = clusters("/async"), 6
	a.refreshSlash()
	if a.slash == nil {
		t.Fatal("no command list")
	}
	found := false
	for _, opt := range a.slash.sel.options {
		if opt.value == "async-tasks" {
			found = true
		}
	}
	if !found {
		t.Error("/async-tasks must not depend on hooks")
	}
}
