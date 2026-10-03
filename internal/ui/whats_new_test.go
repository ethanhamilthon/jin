package ui

import (
	"strings"
	"testing"
)

func TestWhatsNewShowsOncePerUpgrade(t *testing.T) {
	db, _ := openFoldDB(t)
	a := &app{store: db, version: "v0.6"}
	if _, ok := a.whatsNew(); ok {
		t.Fatal("a fresh install must not show release notes")
	}
	_ = db.SetSetting(keyReleaseSeen, "v0.5")
	entry, ok := a.whatsNew()
	if !ok || !strings.HasPrefix(entry.text, "What's new in v0.6") {
		t.Fatalf("upgrade: %v %q", ok, entry.text)
	}
	if _, ok := a.whatsNew(); ok {
		t.Fatal("notes must show only once")
	}
	_ = db.SetSetting(keyReleaseSeen, "")
	_ = db.Touch("s", "/", "m", "", "t")
	if _, ok := a.whatsNew(); !ok {
		t.Fatal("an install from before release.seen existed must see the notes")
	}
}
