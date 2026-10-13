package ui

import (
	"context"
	"testing"
)

func TestTitleFailureStaysNonBlockingForSession(t *testing.T) {
	a, s := persistedApp(t)
	a.sessions[s.id] = s
	a.titles = make(chan titleResult, 1)
	a.receiveTitle(titleResult{id: s.id, err: context.DeadlineExceeded})
	if s.title != "title" || lastText(s) == "" {
		t.Fatalf("title fallback or failure notice lost: %q", lastText(s))
	}
	if _, err := a.store.LoadConfig(); err != nil {
		t.Fatal(err)
	}
}

func TestTitleClaimExistsForAcceptedSession(t *testing.T) {
	a, s := persistedApp(t)
	created, err := a.store.TouchFirst("fresh", s.path, "m", "", "x", "")
	if err != nil || !created {
		t.Fatalf("created = %v, %v", created, err)
	}
	if applied, err := a.store.SetTitle("fresh", "Named"); err != nil || !applied {
		t.Fatalf("applied = %v, %v", applied, err)
	}
}
