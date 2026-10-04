package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

func newFakeSessionsSelector(searchFn func([]string) ([]option, error), labels ...string) (*app, *selector) {
	screen, _ := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 72, Y: 28}))
	_ = screen.Init()
	a := &app{screen: screen, width: 72}
	opts := make([]option, len(labels))
	for i, l := range labels {
		opts[i] = option{label: l, value: l, detail: "detail:" + l}
	}
	sel := &selector{title: "Sessions · /work", options: opts}
	sel.allOptions = append([]option(nil), opts...)
	sel.filter = searchFn
	a.sel = sel
	return a, sel
}

func TestSessionsFilter_LiveTypingFakeSearch(t *testing.T) {
	var lastWords []string
	fakeSearch := func(words []string) ([]option, error) {
		lastWords = words
		var out []option
		for _, w := range words {
			if w == "two" {
				out = append(out, option{label: "Session Two", value: "2"})
			}
		}
		return out, nil
	}
	a, sel := newFakeSessionsSelector(fakeSearch, "Session One", "Session Two")
	typeRune(a, "t")
	typeRune(a, "w")
	typeRune(a, "o")
	if len(lastWords) != 1 || lastWords[0] != "two" {
		t.Fatalf("words = %v, want [two]", lastWords)
	}
	if len(sel.options) != 1 || sel.current() != "2" {
		t.Fatalf("options = %v, current = %q", sel.options, sel.current())
	}
}

func TestSessionsFilter_BackspaceRestoresOriginalOptions(t *testing.T) {
	fakeSearch := func(words []string) ([]option, error) {
		return []option{{label: "Session Two", value: "2"}}, nil
	}
	a, sel := newFakeSessionsSelector(fakeSearch, "1", "2", "3")
	typeRune(a, "t")
	bksp := tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone)
	a.selectorKey(bksp)
	if len(sel.options) != 3 {
		t.Fatalf("options count = %d, want 3", len(sel.options))
	}
}

func TestSessionsFilter_EscClearsFilterFirstLeavesListSecond(t *testing.T) {
	fakeSearch := func(words []string) ([]option, error) {
		return []option{{label: "Session Two", value: "2"}}, nil
	}
	a, sel := newFakeSessionsSelector(fakeSearch, "1", "2")
	typeRune(a, "x")
	esc := tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)
	a.selectorKey(esc)
	if a.sel == nil {
		t.Fatal("first Esc should not close selector")
	}
	if len(sel.query) != 0 || len(sel.options) != 2 {
		t.Fatalf("first Esc should clear filter, options = %d", len(sel.options))
	}
	a.selectorKey(esc)
	if a.sel != nil {
		t.Fatal("second Esc should close selector")
	}
}

func TestSessionsFilter_HeaderDisplaysFilterText(t *testing.T) {
	fakeSearch := func(words []string) ([]option, error) { return nil, nil }
	a, _ := newFakeSessionsSelector(fakeSearch, "1")
	typeRune(a, "b")
	typeRune(a, "u")
	typeRune(a, "g")
	a.drawRuleTitle(1, 72, a.sel)
	if snap := snapshot(a.screen); !strings.Contains(snap, "/bug") {
		t.Fatalf("header missing /bug:\n%s", snap)
	}
}
