package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
	"jin/internal/startup"
	"jin/internal/store"
	"jin/internal/tools"
)

// startingApp is an app with one session that has begun its render.
func startingApp(t *testing.T) *app {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	a, _ := layoutApp(t)
	a.dir = t.TempDir()
	a.ctx = t.Context()
	a.rendered = make(chan renderEvent, 32)
	a.updates = make(chan taggedUpdate, 32)
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	a.store = db
	return a
}

func writeFileT(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

// pump applies render events until the session is ready, or fails on timeout.
func pump(t *testing.T, a *app, s *chatSession) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for !s.ready {
		select {
		case ev := <-a.rendered:
			a.receiveRender(ev)
		case <-deadline:
			t.Fatal("the session never became ready")
		}
	}
}

func newStarting(t *testing.T, a *app) *chatSession {
	t.Helper()
	s := a.startSession("s1", "", "m", "", nil, a.introEntries())
	a.active = s
	return s
}

func TestSessionIsNotReadyUntilItsCommandsFinish(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "slow.md"), "slow {{sleep 0.5; echo done}}")
	s := newStarting(t, a)
	if s.ready || s.render == nil {
		t.Fatal("a session with a slow prompt must start closed")
	}
	pump(t, a, s)
	if s.promptBodies["slow"] != "slow done" {
		t.Errorf("slow = %q", s.promptBodies["slow"])
	}
	if s.render != nil || s.initial != nil {
		t.Error("render state must be cleared when ready")
	}
}

func TestNothingIsTypedOrSentBeforeTheSessionIsReady(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "slow.md"), "{{sleep 0.5}}x")
	s := newStarting(t, a)
	s.input, s.cursor = clusters("draft"), 5
	for _, ch := range []string{"a", "b", "/", "#", "@"} {
		a.handleEvent(tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone))
	}
	a.handleEvent(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	a.handleEvent(tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone))
	a.handleEvent(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone))
	if got := strings.Join(s.input, ""); got != "draft" || s.cursor != 5 {
		t.Errorf("input = %q cursor %d: the closed input changed", got, s.cursor)
	}
	if a.slash != nil || a.mention != nil || a.file != nil || a.sel != nil {
		t.Error("nothing may open while the session starts")
	}
	if len(s.pending) != 0 {
		t.Error("nothing may be queued")
	}
	a.sendDraft("hello")
	if len(s.pending) != 0 || len(s.history) > 0 && s.history[len(s.history)-1].kind == core.UpdateUser {
		t.Error("sendDraft must do nothing before ready")
	}
	box := a.inputBox()
	if box.focused || !strings.Contains(box.placeholder, "Loading") {
		t.Errorf("the closed input box = %+v", box)
	}
	pump(t, a, s)
	a.handleEvent(tcell.NewEventKey(tcell.KeyRune, "z", tcell.ModNone))
	if got := strings.Join(s.input, ""); got != "draftz" {
		t.Errorf("after ready the input must work, got %q", got)
	}
	if !a.inputBox().focused {
		t.Error("the input box must be open when ready")
	}
}

func TestPromptNamesShowASpinnerUntilTheirCommandsFinish(t *testing.T) {
	a := startingApp(t)
	root := filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts")
	writeFileT(t, filepath.Join(root, "fast.md"), "fast {{echo f}}")
	writeFileT(t, filepath.Join(root, "slow.md"), "slow {{sleep 1; echo s}}")
	s := newStarting(t, a)
	plain := func() chatEntry {
		for _, e := range s.history {
			if e.tool == promptsEntry {
				return e
			}
		}
		t.Fatal("no Prompts section")
		return chatEntry{}
	}
	if got := plain().pending; len(got) != 5 {
		t.Fatalf("pending = %v, want every enabled prompt while loading", got)
	}
	// Wait for the fast prompt only.
	deadline := time.After(5 * time.Second)
	for s.isLoading("fast") {
		select {
		case ev := <-a.rendered:
			a.receiveRender(ev)
		case <-deadline:
			t.Fatal("fast never finished")
		}
	}
	pending := plain().pending
	if s.isLoading("fast") || !s.isLoading("slow") {
		t.Fatalf("loading = %v", s.loadingPrompts())
	}
	for _, name := range pending {
		if name == "fast" {
			t.Error("a finished prompt must lose its spinner")
		}
	}
	rows := promptsRows(plain(), 80)
	var spun, names string
	for _, row := range rows {
		for _, span := range row.spans {
			if span.spin {
				spun += "*"
			}
			names += span.text
		}
	}
	if !strings.Contains(names, "#slow") || spun == "" {
		t.Errorf("the slow prompt needs a spinner, got %q (%d spinners)", names, len(spun))
	}
	pump(t, a, s)
	if got := plain().pending; len(got) != 0 {
		t.Errorf("pending after ready = %v", got)
	}
	for _, row := range promptsRows(plain(), 80) {
		for _, span := range row.spans {
			if span.spin {
				t.Error("no spinner may remain when everything is done")
			}
		}
	}
}

func TestCtrlCStopsTheCommandsAndOpensTheSession(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "hang.md"), "hangs {{sleep 30}}")
	s := newStarting(t, a)
	time.Sleep(200 * time.Millisecond)
	start := time.Now()
	a.handleEvent(tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModNone))
	pump(t, a, s)
	if time.Since(start) > 5*time.Second {
		t.Errorf("took %v", time.Since(start))
	}
	if s.promptBodies["hang"] != "hangs [command cancelled]" {
		t.Errorf("hang = %q", s.promptBodies["hang"])
	}
	last := s.history[len(s.history)-1]
	if !strings.Contains(last.text, "Prompt commands cancelled") {
		t.Errorf("last entry = %+v", last)
	}
}

func TestAFailingCommandIsOneLineInTheChat(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "bad.md"), "x {{exit 3}}")
	s := newStarting(t, a)
	pump(t, a, s)
	last := s.history[len(s.history)-1]
	if !strings.Contains(last.text, "Prompt commands:") || !strings.Contains(last.text, "exit status 3") {
		t.Errorf("last entry = %+v", last)
	}
	if s.promptBodies["bad"] != "x [command failed: exit status 3]" {
		t.Errorf("bad = %q", s.promptBodies["bad"])
	}
}

func TestPromptsAreFilledInWhenSent(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "branch.md"), "On branch {{echo main}}.")
	s := newStarting(t, a)
	pump(t, a, s)
	s.agent = core.NewAgent(nil, "", tools.NewRegistry())
	s.sendFiles("fix it #branch", "fix it #branch", "", []string{"branch"})
	if len(s.pending) != 1 || !strings.Contains(s.pending[0].Prompt, "On branch main.") {
		t.Errorf("pending = %+v", s.pending)
	}
	if strings.Contains(s.pending[0].Prompt, "{{") {
		t.Error("the model must get the output, not the command")
	}
}

func TestTaskResultsWaitInTheQueueUntilTheSessionIsReady(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "slow.md"), "{{sleep 0.5}}x")
	s := newStarting(t, a)
	s.sendAsync("<async-task-result id=\"a\" status=\"done\" exit=\"0\">\nok\n</async-task-result>")
	a.flushPending()
	if len(s.pending) != 1 {
		t.Fatalf("pending = %d: a result must wait while the agent is not running", len(s.pending))
	}
	pump(t, a, s)
	a.flushPending()
	if len(s.pending) != 0 {
		t.Errorf("pending = %d after ready: the result must be handed over", len(s.pending))
	}
}

func TestAutocompleteOffersOnlyThePromptsOfTheSession(t *testing.T) {
	a := startingApp(t)
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "deploy.md"), "go")
	a.cfg.PromptsDisabled = []string{"review"}
	s := newStarting(t, a)
	pump(t, a, s)
	s.input, s.cursor = clusters("#"), 1
	a.refreshMention()
	if a.mention == nil {
		t.Fatal("no autocomplete")
	}
	var got []string
	for _, opt := range a.mention.sel.options {
		got = append(got, opt.value)
	}
	want := "deploy plan subagents"
	if strings.Join(got, " ") != want {
		t.Errorf("options = %v, want %s", got, want)
	}
}

var _ = startup.Output{}
var _ = context.Background
