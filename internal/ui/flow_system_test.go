package ui

import (
	"os"
	"strings"
	"testing"

	"jin/internal/sysprompt"
)

func TestSystemPromptSettingCreatesTheFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if _, ok := slashByName("settings"); !ok {
		t.Fatal("/settings is missing")
	}
	a, _ := layoutApp(t)
	a.cfg.Editor = "" // no editor chosen: the flow asks for one, after the file exists
	a.openSettingsFlow()
	a.sel.selectValue("System prompt")
	a.submitSelector()
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

func TestTasksCommandIsAlwaysThere(t *testing.T) {
	a, _ := layoutApp(t)
	a.cfg.HooksDisabled = []string{"async", "docs"}
	a.active.input, a.active.cursor = clusters("/tas"), 4
	a.refreshSlash()
	if a.slash == nil {
		t.Fatal("no command list")
	}
	found := false
	for _, opt := range a.slash.sel.options {
		if opt.value == "tasks" {
			found = true
		}
	}
	if !found {
		t.Error("/tasks must not depend on hooks")
	}
}
