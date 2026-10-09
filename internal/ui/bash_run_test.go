package ui

import (
	"context"
	"strings"
	"testing"
)

func TestShellCommandIsNonInteractive(t *testing.T) {
	dir := t.TempDir()
	cmd := shellCommand(context.Background(), dir, "echo $GIT_TERMINAL_PROMPT $GIT_EDITOR $PAGER $DEBIAN_FRONTEND; cat /dev/tty")
	if cmd.Dir != dir {
		t.Fatalf("command directory = %q, want %q", cmd.Dir, dir)
	}
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "0 true cat noninteractive") {
		t.Fatalf("out = %q, err = %v", out, err)
	}
}

func TestBashResultReturnsToOriginatingSession(t *testing.T) {
	a, _ := layoutApp(t)
	origin, other := &chatSession{id: "origin"}, &chatSession{id: "other"}
	a.sessions = map[string]*chatSession{"origin": origin, "other": other}
	a.receiveBash(bashResult{session: "origin", output: "result"})
	if len(origin.history) != 1 || origin.history[0].text != "result" || len(other.history) != 0 {
		t.Fatalf("origin=%+v other=%+v", origin.history, other.history)
	}
}
