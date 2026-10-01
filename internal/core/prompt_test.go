package core

import (
	"strings"
	"testing"
)

func TestSystemPromptFillsEnvironment(t *testing.T) {
	prompt := SystemPrompt("/work/project")
	if strings.Contains(prompt, "{{") {
		t.Fatalf("unreplaced placeholder in prompt:\n%s", prompt)
	}
	if !strings.Contains(prompt, "- Working directory: /work/project\n") {
		t.Fatalf("working directory missing:\n%s", prompt)
	}
}
