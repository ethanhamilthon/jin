package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
)

func TestSessionsFilter_MessageMatchAndResume(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	dir := "/work/project"
	_ = db.Touch("s1", dir, "m", "high", "Bug Report")
	_ = db.AppendMessage("s1", provider.Message{Role: "user", Content: "Need help with pagination"})

	_ = db.Touch("s2", dir, "m", "high", "Feature Request")
	_ = db.AppendMessage("s2", provider.Message{Role: "user", Content: "Add secret credentials"})

	a, _ := goldenApp(t)
	a.ctx = t.Context()
	a.active.persisted = true
	a.store = db
	a.dir = dir
	a.registry = tools.NewRegistry()

	sel := a.openSessionsFlow()
	if len(sel.options) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sel.options))
	}

	typeRune(a, "p")
	typeRune(a, "a")
	typeRune(a, "g")
	typeRune(a, "i")
	typeRune(a, "n")

	if len(sel.options) != 1 {
		t.Fatalf("expected 1 session matching 'pagin', got %d", len(sel.options))
	}
	if sel.options[0].value != "s1" {
		t.Fatalf("expected session s1, got %s", sel.options[0].value)
	}
	if !sel.matches(0) {
		t.Fatal("session matched on message should pass sel.matches")
	}

	enter := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	a.selectorKey(enter)
	if a.sel != nil {
		t.Fatal("selector should close after enter")
	}
	if a.active.id != "s1" {
		t.Fatalf("active session = %q, want s1", a.active.id)
	}
}

func TestSessionsFilter_PreservesSelection(t *testing.T) {
	fakeSearch := func(words []string) ([]option, error) {
		return []option{{label: "A", value: "A"}, {label: "B", value: "B"}}, nil
	}
	a, sel := newFakeSessionsSelector(fakeSearch, "A", "B", "C")
	sel.index = 1
	typeRune(a, "x")
	if sel.index != 1 || sel.current() != "B" {
		t.Fatalf("expected selection B preserved, index=%d current=%q", sel.index, sel.current())
	}
}
