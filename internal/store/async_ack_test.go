package store

import (
	"os"
	"testing"
)

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
