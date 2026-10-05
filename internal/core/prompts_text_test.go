package core

import (
	"strings"
	"testing"
)

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
