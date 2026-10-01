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
