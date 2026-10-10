package headless

import (
	"testing"

	"jin/internal/core"
	"jin/internal/session"
)

// Only the assistant entries after our own user entry answer our request;
// text another client produced in between is not ours.
func TestAnswerEntriesFollowsOurOwnMessage(t *testing.T) {
	entries := []session.Entry{
		{Kind: core.UpdateUser, Text: "first"},
		{Kind: core.UpdateAssistant, Text: "answer one"},
		{Kind: core.UpdateUser, Text: "ours"},
		{Kind: core.UpdateAssistant, Text: "our answer"},
	}
	got := answerEntries(entries, 1, "ours")
	if len(got) != 1 || got[0].Text != "our answer" {
		t.Fatalf("got %+v", got)
	}
}

// A turn another client started answers that client, so its text is not
// reported as ours.
func TestAnswerEntriesRejectsAnotherClientsTurn(t *testing.T) {
	entries := []session.Entry{
		{Kind: core.UpdateUser, Text: "theirs"},
		{Kind: core.UpdateAssistant, Text: "their answer"},
	}
	if got := answerEntries(entries, 0, "ours"); len(got) != 0 {
		t.Fatalf("another client's answer was reported: %+v", got)
	}
}
