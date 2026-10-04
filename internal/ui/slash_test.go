package ui

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func typeText(a *app, text string) {
	for _, c := range clusters(text) {
		if c == "\n" {
			a.handleEvent(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModShift))
			continue
		}
		a.handleEvent(tcell.NewEventKey(tcell.KeyRune, c, tcell.ModNone))
	}
}

func press(a *app, key tcell.Key) { a.handleEvent(tcell.NewEventKey(key, "", tcell.ModNone)) }

func TestSlashAtOnlyAfterWhitespaceOrStart(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"/mo", true},
		{"fix this /mo", true},
		{"line\n/mo", true},
		{"/usr/bin", false},
		{"and/or", false},
		{"https://x.y/mo", false},
		{"/mo del", false},
	}
	for _, tc := range cases {
		in := clusters(tc.text)
		_, _, ok := slashAt(in, len(in))
		if ok != tc.want {
			t.Errorf("slashAt(%q) = %v, want %v", tc.text, ok, tc.want)
		}
	}
}

func TestSlashListFiltersByPrefix(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "/se")
	if a.slash == nil {
		t.Fatal("typing /se should open the command list")
	}
	var names []string
	for _, o := range a.slash.sel.options {
		names = append(names, o.value)
	}
	if strings.Join(names, ",") != "sessions,settings" {
		t.Fatalf("options = %v", names)
	}
}

func TestSlashInTheMiddleOfTextRunsAndKeepsTheRest(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.input = clusters("before after")
	a.active.cursor = len(clusters("before "))
	typeText(a, "/clear")
	if a.slash == nil {
		t.Fatal("slash list should open in the middle of text")
	}
	// /clear wipes the whole draft by design.
	press(a, tcell.KeyEnter)
	if len(a.active.input) != 0 {
		t.Fatalf("draft = %q", strings.Join(a.active.input, ""))
	}
}

func TestSlashCommandTokenIsCutFromDraft(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.input = clusters("fix this ")
	a.active.cursor = len(a.active.input)
	typeText(a, "/quit")
	press(a, tcell.KeyTab)
	if tok, ok := tokenOf(a.active.input[9]); !ok || tok.payload != "quit" {
		t.Fatalf("Tab on the list must make a token, draft %q", a.active.input)
	}
	press(a, tcell.KeyEnter)
	if got := strings.Join(a.active.input, ""); strings.TrimSpace(got) != "fix this" {
		t.Fatalf("draft = %q, want the token cut with one space", got)
	}
	if a.slash != nil {
		t.Fatal("list should close after running")
	}
	if len(a.active.pending) != 0 {
		t.Fatal("a command token runs, it is not sent")
	}
}

func TestEscClosesSlashListAndKeepsTokenAsText(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "/mo")
	press(a, tcell.KeyEscape)
	if a.slash != nil {
		t.Fatal("Esc should close the list")
	}
	if got := strings.Join(a.active.input, ""); got != "/mo" {
		t.Fatalf("draft = %q", got)
	}
	typeText(a, "d")
	if a.slash == nil {
		t.Fatal("changing the token should reopen the list")
	}
}

func TestSlashListNeverSendsOnEnter(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "/cl")
	press(a, tcell.KeyEnter)
	if len(a.active.pending) != 0 {
		t.Fatal("Enter on the list must run the command, not send the text")
	}
}

func TestNewSessionCarriesTheDraft(t *testing.T) {
	a, _ := layoutApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	a.ctx, a.updates, a.dir = ctx, make(chan taggedUpdate, 8), t.TempDir()
	a.active.stop = cancel
	a.active.input = clusters("keep me")
	a.active.cursor = len(a.active.input)
	old := a.active
	a.newSessionWithDraft()
	if a.active == old {
		t.Fatal("expected a new session")
	}
	if got := strings.Join(a.active.input, ""); got != "keep me" || a.active.cursor != len(a.active.input) {
		t.Fatalf("new draft = %q cursor=%d", got, a.active.cursor)
	}
	if len(old.input) != 0 {
		t.Fatal("the old session should lose the draft")
	}
}

func TestInlineTUIParsesArgsToEndOfLine(t *testing.T) {
	in := slices.Concat(clusters("look "), []string{commandToken("tui")}, clusters(" lazygit -p x\nnext line"))
	cmd, start, end, arg, ok := inlineCommand(in)
	if !ok || cmd.name != "tui" {
		t.Fatal("expected /tui")
	}
	if arg != "lazygit -p x" {
		t.Fatalf("arg = %q", arg)
	}
	if got := strings.Join(in[:start], ""); got != "look " {
		t.Fatalf("before = %q", got)
	}
	if got := strings.Join(in[end:], ""); got != "\nnext line" {
		t.Fatalf("after = %q", got)
	}
}

func TestInlineCommandIgnoresPlainText(t *testing.T) {
	for _, text := range []string{"see /tuilib/x and a/tui b", "/tui lazygit"} {
		if _, _, _, _, ok := inlineCommand(clusters(text)); ok {
			t.Errorf("%q: only a command token runs", text)
		}
	}
}

func TestEveryCommandHasIconAndDescription(t *testing.T) {
	seen := map[string]bool{}
	var names []string
	for _, c := range slashCommands() {
		if c.icon == "" || c.desc == "" || c.run == nil {
			t.Errorf("/%s is incomplete", c.name)
		}
		if seen[c.name] {
			t.Errorf("/%s listed twice", c.name)
		}
		seen[c.name] = true
		names = append(names, c.name)
	}
	want := "sessions projects model provider theme settings reload compact context handoff rewind undo stop new clear edit tui vertical horizontal quit qa tasks"
	if got := strings.Join(names, " "); got != want {
		t.Errorf("commands = %q, want %q", got, want)
	}
	for _, gone := range []string{"voice", "voice-provider", "bash"} {
		if seen[gone] {
			t.Errorf("/%s should not exist", gone)
		}
	}
}

func TestSlashOpensAgainAfterEscAndRetype(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "/")
	press(a, tcell.KeyEscape)
	if a.slash != nil {
		t.Fatal("Esc should close the list")
	}
	press(a, tcell.KeyBackspace2)
	typeText(a, "/")
	if a.slash == nil {
		t.Fatal("a bare / typed again should open the list")
	}
}

func TestBashIsNotAvailableAsASlashCommand(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "keep /ba")
	if a.slash != nil {
		t.Fatal("/bash must not appear in slash completion")
	}
	if _, ok := slashByName("bash"); ok {
		t.Fatal("/bash must not have a hidden command alias")
	}
	if a.active.bash != nil {
		t.Fatal("typing /ba must not start a shell")
	}
}

func TestSlashEnterCompletesCommandWithArguments(t *testing.T) {
	a, _ := layoutApp(t)
	typeText(a, "/tu")
	press(a, tcell.KeyEnter)
	tok, ok := tokenOf(a.active.input[0])
	if !ok || tok.payload != "tui" || a.slash != nil {
		t.Fatalf("draft=%q slash=%v", a.active.input, a.slash)
	}
	if len(a.active.history) != 0 {
		t.Fatal("a command needing arguments must not run yet")
	}
}
