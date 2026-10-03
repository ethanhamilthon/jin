package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestShortcutsWorkInAnyLayout(t *testing.T) {
	cases := []struct {
		name string
		ev   *tcell.EventKey
		want bool
	}{
		{"legacy ctrl+v", tcell.NewEventKey(tcell.KeyCtrlV, "", tcell.ModCtrl), true},
		{"latin ctrl+v", tcell.NewEventKeyEx(tcell.KeyRune, "v", tcell.ModCtrl, true, tcell.KeyV, 1), true},
		{"cyrillic text, physical v", tcell.NewEventKeyEx(tcell.KeyRune, "м", tcell.ModCtrl, true, tcell.KeyV, 1), true},
		{"cyrillic text only", tcell.NewEventKeyEx(tcell.KeyRune, "м", tcell.ModCtrl, true, 0, 1), true},
		{"cmd+v", tcell.NewEventKeyEx(tcell.KeyRune, "v", tcell.ModMeta, true, tcell.KeyV, 1), true},
		{"plain м", tcell.NewEventKeyEx(tcell.KeyRune, "м", tcell.ModNone, true, tcell.KeyV, 1), false},
		{"ctrl+c", tcell.NewEventKeyEx(tcell.KeyRune, "с", tcell.ModCtrl, true, tcell.KeyC, 1), false},
	}
	for _, tc := range cases {
		if got := isPasteKey(tc.ev); got != tc.want {
			t.Errorf("%s: isPasteKey = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCtrlMInCyrillicCyclesNotEnter(t *testing.T) {
	ev := tcell.NewEventKeyEx(tcell.KeyRune, "ь", tcell.ModCtrl, true, 0, 1)
	if !isCycleModelKey(ev) {
		t.Fatal("Ctrl+ь should be Ctrl+M")
	}
	if isCycleModelKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)) {
		t.Fatal("Enter must not cycle the model")
	}
}

func TestPanelActionsInAnyLayout(t *testing.T) {
	ran := ""
	sel := &selector{actions: map[rune]func(string){'a': func(string) { ran = "a" }}}
	sel.actionKey(tcell.NewEventKeyEx(tcell.KeyRune, "ф", tcell.ModNone, true, 0, 1))
	if ran != "a" {
		t.Fatal("ф should run the a action")
	}
}

func TestLayoutTablesLineUp(t *testing.T) {
	for _, layout := range layouts {
		if a, b := len([]rune(layout[0])), len([]rune(layout[1])); a != b {
			t.Errorf("%q has %d letters for %d keys", layout[0], a, b)
		}
	}
	for typed, want := range map[string]rune{"і": 's', "ї": ']', "є": '\'', "ў": 'o', "ψ": 'c', "ς": 'w', "Ζ": 'z', "ά": 'a', "ύ": 'y', "қ": '0'} {
		ev := tcell.NewEventKeyEx(tcell.KeyRune, typed, tcell.ModCtrl, true, 0, 1)
		if got := keyLetter(ev); got != want {
			t.Errorf("keyLetter(%s) = %q, want %q", typed, got, want)
		}
	}
}
