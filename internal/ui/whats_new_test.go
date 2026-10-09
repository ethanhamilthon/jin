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

func TestReleaseNotesForCurrentPatch(t *testing.T) {
	if notes := releaseNotes["v0.9.6"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "paths") {
		t.Fatalf("v0.9.6 notes = %q", notes)
	}
	if notes := releaseNotes["v0.9.5"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "Ctrl+T") {
		t.Fatalf("v0.9.5 notes = %q", notes)
	}
	if notes := releaseNotes["v0.9.4"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "web.accent") {
		t.Fatalf("v0.9.4 notes = %q", notes)
	}
	if notes := releaseNotes["v0.9.3"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "paperclip") {
		t.Fatalf("v0.9.3 notes = %q", notes)
	}
	if notes := releaseNotes["v0.9.2"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "ask_user") {
		t.Fatalf("v0.9.2 notes = %q", notes)
	}
	if notes := releaseNotes["v0.9.1"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "Ctrl+O") {
		t.Fatalf("v0.9.1 notes = %q", notes)
	}
	if notes := releaseNotes["v0.9.0"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "jin web") {
		t.Fatalf("v0.9.0 notes = %q", notes)
	}
	if notes := releaseNotes["v0.8.4"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "task tool") {
		t.Fatalf("v0.8.4 notes = %q", notes)
	}
	if notes := releaseNotes["v0.8.3"]; !strings.Contains(notes, "green") {
		t.Fatalf("v0.8.3 notes = %q", notes)
	}
	if notes := releaseNotes["v0.8.2"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "Esc") {
		t.Fatalf("v0.8.2 notes = %q", notes)
	}
	if notes := releaseNotes["v0.8.1"]; !strings.Contains(notes, "pictures") {
		t.Fatalf("v0.8.1 notes = %q", notes)
	}
	if notes := releaseNotes["v0.8.0"]; len(strings.Split(notes, "\n")) != 5 || !strings.Contains(notes, "four panes") {
		t.Fatalf("v0.8.0 notes = %q", notes)
	}
	if notes := releaseNotes["v0.7.1"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "/voice") {
		t.Fatalf("v0.7.1 notes = %q", notes)
	}
	if notes := releaseNotes["v0.7.0"]; len(strings.Split(notes, "\n")) != 3 || !strings.Contains(notes, "padding") {
		t.Fatalf("v0.7.0 notes = %q", notes)
	}
	if notes := releaseNotes["v0.6.9"]; len(strings.Split(notes, "\n")) != 5 || !strings.Contains(notes, "Long sessions") {
		t.Fatalf("v0.6.9 notes = %q", notes)
	}
	if notes := releaseNotes["v0.6.3"]; !strings.Contains(notes, "Input glow") {
		t.Fatalf("v0.6.3 notes = %q", notes)
	}
	for _, v := range []string{"v0.6.1", "v0.6.2"} {
		if notes := releaseNotes[v]; len(strings.Split(notes, "\n")) != 4 {
			t.Fatalf("%s notes = %q", v, notes)
		}
	}
}
