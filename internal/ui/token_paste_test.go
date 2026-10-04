package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestKeysRightAfterALongPasteSeeTheToken(t *testing.T) {
	for _, pasted := range []string{"a\nb\n#plan", "a\nb\n/tui"} {
		for _, key := range []tcell.Key{tcell.KeyEnter, tcell.KeyTab, tcell.KeyEscape} {
			a, _ := fileApp(t)
			s := a.active
			s.promptBodies = map[string]string{"plan": "PLAN BODY"}
			pasteText(a, pasted)
			if a.mention != nil || a.slash != nil || a.file != nil {
				t.Fatalf("%q: a list stays open after the paste", pasted)
			}
			press(a, key)
			if key == tcell.KeyEnter && (len(s.pending) != 1 || s.pending[0].Prompt != pasted) {
				t.Errorf("%q: Enter must send the paste literally, pending %+v", pasted, s.pending)
			}
			if key != tcell.KeyEnter && (len(s.input) != 1 || shownDraft(s) != "[pasted text 3 lines]") {
				t.Errorf("%q key %v: draft = %q", pasted, key, shownDraft(s))
			}
		}
	}
}
