package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestBashInputUsesRenderedDraft(t *testing.T) {
	for _, tc := range []struct {
		text string
		want bool
	}{
		{"$", true}, {"$ echo", true}, {"x $ echo", false}, {"", false},
	} {
		s := &chatSession{input: clusters(tc.text)}
		if got := s.bashInput(); got != tc.want {
			t.Errorf("bashInput(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
	pasted := pasteToken("$ echo '$HOME'\nnext")
	if !(&chatSession{input: []string{pasted}}).bashInput() {
		t.Fatal("a leading dollar inside a paste token must enable shell mode")
	}
	if (&chatSession{input: []string{imageToken("$image.png")}}).bashInput() {
		t.Fatal("an image path beginning with a dollar is not a leading draft dollar")
	}
}

func TestBashCommandStripsOnlyTheLeadingDollar(t *testing.T) {
	text := ` printf '%s' "$HOME"` + "\necho next"
	got, ok := bashCommand(append(clusters("$"), clusters(text)...))
	if !ok || got != text {
		t.Fatalf("bashCommand = %q, %v; want %q", got, ok, text)
	}
	for _, input := range [][]string{clusters("$"), clusters("$ \n\t")} {
		if command, ok := bashCommand(input); ok || command != "" {
			t.Errorf("empty shell draft returned %q, %v", command, ok)
		}
	}
}

func TestBracketedPasteMarkerDoesNotHideShellMode(t *testing.T) {
	a, _ := layoutApp(t)
	text := "$ printf '%s' \"$HOME\"\necho /quit\nexit"
	a.beginPaste()
	insertClusters(&a.active.input, &a.active.cursor, text)
	a.endPaste()
	if !a.active.bashInput() || draftPayload(a.active.input) != text {
		t.Fatalf("draft payload = %q; shell=%v", draftPayload(a.active.input), a.active.bashInput())
	}
}

func TestShellPasteTokensSuppressAndRestoreAutocomplete(t *testing.T) {
	a, _ := layoutApp(t)
	insertPaste(&a.active.input, &a.active.cursor, "$ /cl\n"+strings.Repeat("literal ", 50))
	if !a.active.bashInput() {
		t.Fatal("long clipboard paste should retain leading shell marker")
	}
	a.active.input, a.active.cursor = clusters("$ /cl"), len(clusters("$ /cl"))
	a.refreshPanels()
	if a.slash != nil || a.mention != nil || a.file != nil {
		t.Fatal("bash drafts must not open Jin completion panels")
	}
	a.active.input, a.active.cursor = clusters(" /cl"), len(clusters(" /cl"))
	a.refreshPanels()
	if a.slash == nil {
		t.Fatal("slash completion should return after removing the leading dollar")
	}
}

func TestShellDraftNeverRunsSlashTokensAsJinCommands(t *testing.T) {
	a, _ := layoutApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	a.ctx = ctx
	a.active.input = append(clusters("$ "), commandToken("quit"))
	a.active.cursor = len(a.active.input)
	before := draftPayload(a.active.input)
	a.tokenizeCommand(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone))
	if draftPayload(a.active.input) != before {
		t.Fatal("shell draft was tokenized as a Jin command")
	}
	a.insertKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if a.quit || len(a.active.pending) != 0 || a.active.bash == nil || !a.active.bash.running {
		t.Fatalf("shell submit ran as Jin command: quit=%v pending=%d bash=%+v", a.quit, len(a.active.pending), a.active.bash)
	}
}
