package core

import (
	"strings"
	"testing"
)

func TestMacOSNoticeOnlyOnDarwin(t *testing.T) {
	darwin := BuildSystemPrompt(PromptInput{System: "x", OS: "darwin", ToolNames: []string{"edit"}})
	if !strings.Contains(darwin, "never run sed -i") || !strings.Contains(darwin, "perl -pi -e") {
		t.Errorf("darwin prompt lacks the sed notice:\n%s", darwin)
	}
	linux := BuildSystemPrompt(PromptInput{System: "x", OS: "linux", ToolNames: []string{"edit"}})
	if strings.Contains(linux, "sed -i") {
		t.Errorf("linux prompt has the macOS notice:\n%s", linux)
	}
}

func TestNoticesSitBeforeTheCacheBreak(t *testing.T) {
	prompt := BuildSystemPrompt(PromptInput{System: "x", OS: "darwin", ToolNames: []string{"write"}})
	cut := strings.Index(prompt, "Environment:")
	if cut < 0 || strings.Index(prompt, "sed -i") > cut || strings.Index(prompt, "TODO.md") > cut {
		t.Errorf("notices must come before the environment block:\n%s", prompt)
	}
}

func TestTodoFileNoticeNeedsAFileTool(t *testing.T) {
	for tools, want := range map[string]bool{"write": true, "edit": true, "read,bash": false, "": false} {
		var names []string
		if tools != "" {
			names = strings.Split(tools, ",")
		}
		got := strings.Contains(BuildSystemPrompt(PromptInput{System: "x", OS: "linux", ToolNames: names}), "TODO.md")
		if got != want {
			t.Errorf("tools %q: TODO.md notice = %v, want %v", tools, got, want)
		}
	}
}
