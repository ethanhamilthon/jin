package ui

import (
	"context"
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

func TestSentTokensAreForgotten(t *testing.T) {
	a, _ := fileApp(t)
	other := &chatSession{input: []string{pasteToken("kept\n2\n3")}}
	a.sessions["other"] = other
	pasteText(a, "1\n2\n3")
	sent := a.active.input[0]
	press(a, tcell.KeyEnter)
	if _, ok := tokenOf(sent); ok {
		t.Error("a sent token must leave the registry")
	}
	if _, ok := tokenOf(other.input[0]); !ok {
		t.Error("a token in another draft must stay")
	}
}

func TestNewSessionCarriesDraftTokens(t *testing.T) {
	a, _ := layoutApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	a.ctx, a.updates, a.dir = ctx, make(chan taggedUpdate, 8), t.TempDir()
	a.active.stop = cancel
	a.active.input = []string{pasteToken("1\n2\n3"), imageToken("/tmp/a.png")}
	a.active.cursor = 2
	a.newSessionWithDraft()
	a.pruneTokens()
	if got := draftPayload(a.active.input); got != "1\n2\n3/tmp/a.png" {
		t.Errorf("moved draft = %q", got)
	}
}
