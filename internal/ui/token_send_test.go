package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func pasteText(a *app, text string) {
	a.handleEvent(tcell.NewEventPaste(true))
	for _, c := range clusters(text) {
		if c == "\n" {
			press(a, tcell.KeyEnter)
			continue
		}
		a.handleEvent(tcell.NewEventKey(tcell.KeyRune, c, tcell.ModNone))
	}
	a.handleEvent(tcell.NewEventPaste(false))
}

func TestLongPasteIsSentLiterally(t *testing.T) {
	a, _ := fileApp(t)
	s := a.active
	s.promptBodies = map[string]string{"plan": "PLAN BODY"}
	pasted := "see #plan\n/undo now\nread @notes.md {{echo hi}}"
	pasteText(a, pasted)
	if len(s.input) != 1 || shownDraft(s) != "[pasted text 3 lines]" {
		t.Fatalf("draft = %q", shownDraft(s))
	}
	typeText(a, " ok")
	press(a, tcell.KeyEnter)
	if len(s.pending) != 1 {
		t.Fatalf("pending = %d", len(s.pending))
	}
	if got := s.pending[0].Prompt; got != pasted+" ok" {
		t.Errorf("prompt = %q", got)
	}
	if last := s.history[len(s.history)-1]; last.text != "[pasted text 3 lines] ok" {
		t.Errorf("chat shows %q", last.text)
	}
}

func TestShortPasteIsPlainText(t *testing.T) {
	a, _ := layoutApp(t)
	pasteText(a, "one\ntwo")
	if got := strings.Join(a.active.input, ""); got != "one\ntwo" || a.active.cursor != 7 {
		t.Fatalf("input = %q cursor %d", got, a.active.cursor)
	}
	long := strings.Repeat("x", 301)
	var input []string
	cursor := 0
	insertPaste(&input, &cursor, long)
	if len(input) != 1 || shownCluster(input[0]) != "[pasted text 1 line]" {
		t.Errorf("a paste over 300 characters must be a token, got %d elements", len(input))
	}
}

func TestImageTokensAreNumberedInOrder(t *testing.T) {
	a, _ := fileApp(t)
	s := a.active
	s.input = slices.Concat(clusters("see "), []string{imageToken("/tmp/a.png")}, clusters(" and "), []string{imageToken("/tmp/b.png")})
	s.cursor = len(s.input)
	press(a, tcell.KeyEnter)
	if last := s.history[len(s.history)-1]; last.text != "see [image 01] and [image 02]" {
		t.Errorf("chat shows %q", last.text)
	}
	if got := s.pending[0].Prompt; got != "see [image 01: /tmp/a.png] and [image 02: /tmp/b.png]" {
		t.Errorf("prompt = %q", got)
	}
}

func TestOnlyAcceptedPromptsExpand(t *testing.T) {
	a, _ := fileApp(t)
	s := a.active
	s.promptBodies = map[string]string{"plan": "PLAN BODY"}
	typeText(a, "#pl")
	press(a, tcell.KeyEnter)
	typeText(a, "then #plan x")
	press(a, tcell.KeyEnter)
	if len(s.pending) != 1 {
		t.Fatalf("pending = %d, draft %q", len(s.pending), shownDraft(s))
	}
	prompt := s.pending[0].Prompt
	if strings.Count(prompt, "PLAN BODY") != 1 || !strings.HasSuffix(prompt, "#plan then #plan x") {
		t.Errorf("prompt = %q", prompt)
	}
}
