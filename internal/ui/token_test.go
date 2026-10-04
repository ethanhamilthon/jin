package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func shownDraft(s *chatSession) string {
	return renderTokens(strings.Join(s.input, ""), tokenLabel)
}

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

func TestBackspaceDeletesATokenWhole(t *testing.T) {
	a, _ := layoutApp(t)
	s := a.active
	s.input = slices.Concat(clusters("ab"), []string{pasteToken("x\ny\nz")}, clusters("cd"))
	s.cursor = 3
	press(a, tcell.KeyBackspace2)
	if got := strings.Join(s.input, ""); got != "abcd" || s.cursor != 2 {
		t.Fatalf("input = %q cursor %d", got, s.cursor)
	}
}

func TestCursorStepsOverATokenAndDrawsAfterItsLabel(t *testing.T) {
	a, screen := layoutApp(t)
	s := a.active
	s.input = slices.Concat(clusters("ab"), []string{imageToken("/tmp/a.png")}, clusters("c"))
	s.cursor = 2
	press(a, tcell.KeyRight)
	if s.cursor != 3 {
		t.Fatalf("cursor = %d, want 3: a token is one step", s.cursor)
	}
	if _, _, col := wrapInput(s.input, s.cursor, 50); col != 2+len("[image 01]") {
		t.Errorf("cursor column = %d", col)
	}
	a.draw()
	found := false
	for y := 0; y < 24; y++ {
		found = found || strings.Contains(rowText(screen, y, 60), "ab[image 01]c")
	}
	if !found {
		t.Error("the input must show the token label")
	}
}

func TestWrapKeepsATokenOnOneLine(t *testing.T) {
	input := slices.Concat(clusters("hello "), []string{pasteToken("1\n2\n3")})
	lines, row, col := wrapInput(input, len(input), 25)
	if len(lines) != 2 || len(lines[1]) != 1 || !isTokenMark(lines[1][0]) {
		t.Fatalf("lines = %q", lines)
	}
	if row != 1 || col != len("[pasted text 3 lines]") {
		t.Errorf("cursor = %d,%d", row, col)
	}
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

func TestTypedCommandWithSpaceBecomesAToken(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "/tui lazygit")
	s := a.active
	if tok, ok := tokenOf(s.input[0]); !ok || tok.kind != tokenCommand || tok.payload != "tui" {
		t.Fatalf("input = %q", s.input)
	}
	if cmd, _, _, arg, ok := inlineCommand(s.input); !ok || cmd.name != "tui" || arg != "lazygit" {
		t.Errorf("inline command = %v %q %v", cmd.name, arg, ok)
	}
	typeText(a, " /undo ")
	if isTokenMark(s.input[len(s.input)-6]) {
		t.Error("a command without arguments stays text")
	}
}

func TestDraftPayloadGivesTokensAsText(t *testing.T) {
	input := slices.Concat([]string{promptToken("plan")}, clusters(" "), []string{pasteToken("a\nb\nc"), imageToken("/tmp/a.png")})
	if got := draftPayload(input); got != "#plan a\nb\nc/tmp/a.png" {
		t.Errorf("payload = %q", got)
	}
}
