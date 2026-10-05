package ui

import (
	"os"
	"strings"
	"testing"
)

func lastText(s *chatSession) string { return s.history[len(s.history)-1].text }

func TestSendIsRefusedWithoutAnyMutationWhenAnotherProcessTookTheSession(t *testing.T) {
	a, s := persistedApp(t)
	a.active, s.ready = s, true
	before, _, _ := a.store.GetSession("s1")
	claimAs(t, os.Getppid(), "s1")
	entries := len(s.history)
	a.sendDraft("hello")
	if len(s.pending) != 0 || s.readOnlyPID != os.Getppid() {
		t.Fatalf("pending %d, readOnlyPID %d", len(s.pending), s.readOnlyPID)
	}
	if string(s.input[0]) != "h" || len(s.input) != 5 {
		t.Errorf("the draft was not given back: %v", s.input)
	}
	if len(s.history) != entries+1 || !strings.HasPrefix(lastText(s), "Read-only: in use by process") {
		t.Errorf("entries: %v", lastText(s))
	}
	after, _, _ := a.store.GetSession("s1")
	if !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Error("session metadata changed")
	}
	if msgs, _ := a.store.LoadMessages("s1"); len(msgs) != 0 {
		t.Errorf("messages = %d", len(msgs))
	}
	a.sendDraft("again")
	if len(s.pending) != 0 {
		t.Error("a read-only session queued a request")
	}
}

func TestSendInADeletedProviderSessionQueuesNothing(t *testing.T) {
	a, s := persistedApp(t)
	a.active, s.ready, s.providerMissing = s, true, true
	a.sendDraft("hello")
	if len(s.pending) != 0 || !strings.Contains(lastText(s), "was deleted") {
		t.Errorf("pending %d, last %q", len(s.pending), lastText(s))
	}
}

func TestCompactReadinessFollowsTheSessionProvider(t *testing.T) {
	a, s := persistedApp(t)
	a.active, s.persisted = s, true
	a.cfg.Provider.APIKey = ""
	s.client.Configure(readyConfig().Provider)
	if err := a.sideRefusal(s); err != nil {
		t.Errorf("a ready session was refused: %v", err)
	}
	a.cfg.Provider = readyConfig().Provider
	s.providerMissing = true
	if err := a.sideRefusal(s); err == nil {
		t.Error("a deleted-provider session may compact")
	}
}
