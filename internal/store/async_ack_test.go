package store

import (
	"os"
	"testing"

	"jin/internal/provider"
)

func TestHasUserMessageMatchesExactText(t *testing.T) {
	db := openTest(t)
	if err := db.AppendMessage("s1", provider.Message{Role: "user", Content: "result \"1\"\n<x>"}); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		session, text string
		want          bool
	}{
		{"s1", "result \"1\"\n<x>", true},
		{"s1", "result", false},
		{"s2", "result \"1\"\n<x>", false},
	}
	for _, c := range cases {
		if got, err := db.HasUserMessage(c.session, c.text); err != nil || got != c.want {
			t.Errorf("HasUserMessage(%q, %q) = %v, %v", c.session, c.text, got, err)
		}
	}
}

func TestReleaseAsyncEventsForOnlyFreesOwnClaims(t *testing.T) {
	db := openTest(t)
	_ = db.AddAsyncEvent("s1", "/p", "one")
	_ = db.AddAsyncEvent("s2", "/p", "two")
	if events, _ := db.ClaimAsyncEvents("/p", os.Getpid()); len(events) != 2 {
		t.Fatalf("claimed %d", len(events))
	}
	if err := db.ReleaseAsyncEventsFor("s1", os.Getpid()); err != nil {
		t.Fatal(err)
	}
	events, _ := db.ClaimAsyncEvents("/p", os.Getpid())
	if len(events) != 1 || events[0].SessionID != "s1" {
		t.Errorf("events = %+v", events)
	}
}
