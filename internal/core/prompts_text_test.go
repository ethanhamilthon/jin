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
