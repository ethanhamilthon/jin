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
	want := `Use the session id from "Your session id: <id>" at the end of this prompt`
	if !strings.Contains(prompt, want) {
		t.Errorf("async prompt does not describe session id location: %s", prompt)
	}
	if !strings.Contains(prompt, "\nYour session id: test-sess-123") {
		t.Errorf("session id line missing: %s", prompt)
	}
}

func TestSummaryLeadContinuationBrief(t *testing.T) {
	msg := SummaryMessage("test summary")
	if !strings.Contains(msg.Content, "continuation brief") {
		t.Errorf("SummaryMessage should mention continuation brief: %s", msg.Content)
	}
	if !strings.Contains(msg.Content, "separate explicit user requirements from observations and unverified claims") {
		t.Errorf("SummaryMessage should separate requirements from unverified claims: %s", msg.Content)
	}
}

func TestSummaryLeadPlanCondition(t *testing.T) {
	msg := SummaryMessage("test summary")
	want := "If it lists unfinished work and the user's next message does not change the plan, continue it."
	if !strings.Contains(msg.Content, want) {
		t.Errorf("SummaryMessage missing conditional continue: %s", msg.Content)
	}
}

func TestDocsPromptCompact(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(docsPrompt), "\n")
	if len(lines) > 4 {
		t.Errorf("docsPrompt should be around 3 lines, got %d lines", len(lines))
	}
	for _, want := range []string{
		"Jin documentation:",
		"README.md` first",
		"only matching files",
		"exact names",
		"never print the API key",
	} {
		if !strings.Contains(docsPrompt, want) {
			t.Errorf("docsPrompt missing %q", want)
		}
	}
}

func TestAsyncPromptIsShortAndPointsToHelp(t *testing.T) {
	if len(asyncPrompt) > 1500 {
		t.Errorf("async prompt grew to %d bytes", len(asyncPrompt))
	}
	for _, want := range []string{"--stdin", "stdin is closed by default", "not from the user", "<async-task-result id=", "jin --help"} {
		if !strings.Contains(strings.ToLower(asyncPrompt), strings.ToLower(want)) {
			t.Errorf("async prompt lacks %q", want)
		}
	}
	for _, gone := range []string{"jin async input --id", "jin async stop --id", "jin async check --id"} {
		if strings.Contains(asyncPrompt, gone) {
			t.Errorf("async prompt still has the usage %q", gone)
		}
	}
}
