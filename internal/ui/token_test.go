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

func TestDraftPayloadGivesTokensAsText(t *testing.T) {
	input := slices.Concat([]string{promptToken("plan")}, clusters(" "), []string{pasteToken("a\nb\nc"), imageToken("/tmp/a.png")})
	if got := draftPayload(input); got != "#plan a\nb\nc/tmp/a.png" {
		t.Errorf("payload = %q", got)
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
}
