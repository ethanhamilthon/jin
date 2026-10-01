package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func testSelector(labels ...string) *selector {
	sel := &selector{}
	for _, label := range labels {
		sel.options = append(sel.options, option{label: label, value: label})
	}
	return sel
}

func TestMoveWrapsAround(t *testing.T) {
	sel := testSelector("a", "b", "c")
	sel.move(-1)
	if sel.index != 2 {
		t.Errorf("Up from the first option should land on the last, got %d", sel.index)
	}
	sel.move(1)
	if sel.index != 0 {
		t.Errorf("Down from the last option should land on the first, got %d", sel.index)
	}
}

func TestMoveWrapsWithinFilter(t *testing.T) {
	sel := testSelector("apple", "banana", "avocado", "cherry")
	sel.query = clusters("a")
	sel.index = 2
	sel.move(1)
	if sel.index != 0 {
		t.Errorf("wrap should skip filtered-out options, got %d", sel.index)
	}
	sel.move(-1)
	if sel.index != 2 {
		t.Errorf("wrap backwards should skip filtered-out options, got %d", sel.index)
	}
}

func TestMoveRecoversWhenSelectionIsFilteredOut(t *testing.T) {
	sel := testSelector("apple", "banana", "avocado")
	sel.query = clusters("av")
	sel.index = 1
	sel.move(-1)
	if sel.index != 2 {
		t.Errorf("got %d, want the last visible option", sel.index)
	}
}

func actionSelector(pressed *[]string) *selector {
	sel := testSelector("alpha", "beta")
	sel.actions = map[rune]func(string){
		'a': func(value string) { *pressed = append(*pressed, "a:"+value) },
		'd': func(value string) { *pressed = append(*pressed, "d:"+value) },
	}
	return sel
}

func typeRune(a *app, r string) {
	a.selectorKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
}

func TestPlainLettersRunActionsUntilSearchOpens(t *testing.T) {
	var pressed []string
	a := &app{sel: actionSelector(&pressed)}
	typeRune(a, "d")
	typeRune(a, "x")
	if len(pressed) != 1 || pressed[0] != "d:alpha" || len(a.sel.query) != 0 {
		t.Fatalf("pressed %v, query %v", pressed, a.sel.query)
	}
	typeRune(a, "/")
	typeRune(a, "b")
	if a.sel.current() != "beta" {
		t.Errorf("search should filter, current %q", a.sel.current())
	}
	typeRune(a, "d")
	if len(pressed) != 1 || len(a.sel.query) != 2 {
		t.Errorf("search should take letters: pressed %v, query %v", pressed, a.sel.query)
	}
}

func TestEscapeClosesSearchBeforePanel(t *testing.T) {
	var pressed []string
	a := &app{sel: actionSelector(&pressed)}
	typeRune(a, "/")
	typeRune(a, "b")
	esc := tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)
	a.selectorKey(esc)
	if a.sel == nil || a.sel.search || len(a.sel.query) != 0 {
		t.Fatalf("first Esc should close only the search: %+v", a.sel)
	}
	a.selectorKey(esc)
	if a.sel != nil {
		t.Error("second Esc should close the panel")
	}
}

func TestListWithoutActionsSearchesImmediately(t *testing.T) {
	a := &app{sel: testSelector("alpha", "beta")}
	typeRune(a, "b")
	if len(a.sel.query) != 1 || a.sel.current() != "beta" {
		t.Errorf("query %v, current %q", a.sel.query, a.sel.current())
	}
	a.selectorKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if a.sel != nil {
		t.Error("Esc should close a list without actions")
	}
}

func choiceSelector(saved *[]string) *selector {
	sel := testSelector("x")
	sel.options = []option{{label: "Volume", value: "volume", choices: []string{"10%", "50%", "100%"}, chosen: 1}}
	sel.onChoice = func(row string, chosen int) error {
		*saved = append(*saved, row+":"+sel.options[0].choices[chosen])
		return nil
	}
	return sel
}

func TestShiftChangesTheFocusedRowAndStopsAtTheEnds(t *testing.T) {
	var saved []string
	sel := choiceSelector(&saved)
	sel.shift(1)
	sel.shift(1)
	sel.shift(-1)
	sel.shift(-1)
	sel.shift(-1)
	want := []string{"volume:100%", "volume:50%", "volume:10%"}
	if len(saved) != len(want) {
		t.Fatalf("saved %v, want %v", saved, want)
	}
	for i := range want {
		if saved[i] != want[i] {
			t.Errorf("change %d = %s, want %s", i, saved[i], want[i])
		}
	}
}

func TestSessionDots(t *testing.T) {
	active := &chatSession{}
	working := &chatSession{working: true}
	unread := &chatSession{unread: true}
	a := &app{active: active, sessions: map[string]*chatSession{"a": active, "w": working, "u": unread}, unread: map[string]bool{"d": true}}
	cases := []struct {
		id    string
		frame int
		dot   string
		style tcell.Style
	}{
		{"a", 0, "●", dotGreen},
		{"w", 0, "●", dotBlue},
		{"w", 6, " ", dotBlue},
		{"u", 6, "●", dotBlue},
		{"d", 6, "●", dotBlue},
		{"none", 0, " ", dotBlue},
	}
	for _, c := range cases {
		a.frame = c.frame
		if dot, style := a.optionMark(c.id); dot != c.dot || style != c.style {
			t.Errorf("%s at frame %d: %q, want %q", c.id, c.frame, dot, c.dot)
		}
	}
}
