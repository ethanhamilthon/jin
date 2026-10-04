package core

import (
	"strings"
	"testing"
)

func TestAsyncPromptReferencesSessionID(t *testing.T) {
	prompt := BuildSystemPrompt(PromptInput{
		System:    "x",
		SessionID: "test-sess-123",
		ToolNames: []string{"bash"},
	})
	if strings.Contains(prompt, "Environment section") {
		t.Error("async prompt should not reference Environment section")
	}
	want := `use the session id from "Your session id: <id>" at the end of the async block`
	if !strings.Contains(prompt, want) {
		t.Errorf("async prompt does not describe session id location: %s", prompt)
	}
	if !strings.Contains(prompt, "- Your session id: test-sess-123") {
		t.Errorf("session id line missing: %s", prompt)
	}
}
