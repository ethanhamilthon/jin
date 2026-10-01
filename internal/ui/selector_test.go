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

func TestCtrlLetter(t *testing.T) {
	cases := []struct {
		name string
		ev   *tcell.EventKey
		want rune
		ok   bool
	}{
		{"legacy ctrl+a", tcell.NewEventKey(tcell.KeyCtrlA, "", tcell.ModNone), 'a', true},
		{"kitty ctrl+d", tcell.NewEventKeyEx(tcell.KeyRune, "d", tcell.ModCtrl, true, 0, 1), 'd', true},
		{"plain letter", tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone), 0, false},
		{"enter", tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone), 0, false},
	}
	for _, c := range cases {
		if got, ok := ctrlLetter(c.ev); got != c.want || ok != c.ok {
			t.Errorf("%s: got %q %v, want %q %v", c.name, got, ok, c.want, c.ok)
		}
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
