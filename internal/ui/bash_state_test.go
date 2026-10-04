package ui

import (
	"context"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestTypingAndDeletingLeadingDollarTogglesShellMode(t *testing.T) {
	for _, key := range []tcell.Key{tcell.KeyBackspace2, tcell.KeyDelete} {
		a, _ := layoutApp(t)
		a.insertKey(tcell.NewEventKey(tcell.KeyRune, "$", tcell.ModNone))
		if !a.active.bashInput() {
			t.Fatal("typing the first dollar did not enable shell mode")
		}
		if key == tcell.KeyDelete {
			a.insertKey(tcell.NewEventKey(tcell.KeyHome, "", tcell.ModNone))
		}
		a.insertKey(tcell.NewEventKey(key, "", tcell.ModNone))
		if a.active.bashInput() {
			t.Fatalf("key %v did not restore normal input", key)
		}
	}
}

func TestRunningShellDoesNotReplaceNormalDraftStatus(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.bash = &bashState{running: true}
	a.insertKey(tcell.NewEventKey(tcell.KeyRune, "m", tcell.ModNone))
	if got := draftPayload(a.active.input); got != "m" {
		t.Fatalf("normal message draft while shell runs = %q", got)
	}
	box := a.inputBox()
	if box.prefix == "$ " || box.placeholder != "Message..." {
		t.Fatalf("normal draft box = %+v", box)
	}
	a.active.input, a.active.cursor = clusters("$ echo"), 6
	box = a.inputBox()
	if box.prefix != "❯ " || box.placeholder != "Shell command · Enter run · Ctrl+C stop" || draftPayload(box.text) != "$ echo" {
		t.Fatalf("shell draft box = %+v", box)
	}
}

func TestEmptyBashDraftDoesNotRun(t *testing.T) {
	a, _ := layoutApp(t)
	for _, text := range []string{"$", "$ \n\t"} {
		a.active.input = clusters(text)
		a.active.cursor = len(a.active.input)
		a.submitBash(a.active)
		if a.active.bash != nil || len(a.active.history) != 0 {
			t.Fatalf("empty draft %q started a command", text)
		}
	}
}

func TestRunningBashRejectsSecondCommandWithoutLosingDraft(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.bash = &bashState{running: true}
	a.active.input = clusters("$ echo second")
	before := draftPayload(a.active.input)
	a.submitBash(a.active)
	if draftPayload(a.active.input) != before || !a.active.bash.running {
		t.Fatal("a second command changed the draft or running state")
	}
	if len(a.active.history) != 1 || a.active.history[0].text != "A shell command is already running." {
		t.Fatalf("history = %+v", a.active.history)
	}
}

func TestReadOnlySessionRejectsShellCommand(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.readOnlyPID = 42
	a.active.input = clusters("$ echo forbidden")
	a.submitBash(a.active)
	if a.active.bash != nil || len(a.active.history) != 1 || a.active.history[0].text != readOnlyText(42) {
		t.Fatalf("bash=%+v history=%+v", a.active.bash, a.active.history)
	}
}

func TestInterruptCancelsFocusedShell(t *testing.T) {
	a, _ := layoutApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.active.bash = &bashState{running: true, cancel: cancel}
	a.interrupt()
	if ctx.Err() == nil {
		t.Fatal("Ctrl+C did not cancel the focused shell")
	}
}
