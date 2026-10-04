package ui

import (
	"os"
	"os/exec"
	"strconv"
	"testing"

	"jin/internal/core"
	"jin/internal/paths"
	"jin/internal/provider"
	"jin/internal/store"
)

func claimAs(t *testing.T, pid int, session string) {
	t.Helper()
	file, err := paths.Global("jin.db")
	if err != nil {
		t.Fatal(err)
	}
	query := "INSERT OR REPLACE INTO running_sessions(session_id, pid) VALUES ('" + session + "', " + strconv.Itoa(pid) + ")"
	if out, err := exec.Command("sqlite3", file, query).CombinedOutput(); err != nil {
		t.Skipf("sqlite3 is not available: %v %s", err, out)
	}
}

func sessionWithOpenToolCall(t *testing.T, a *app, id string) store.Session {
	t.Helper()
	if err := a.store.TouchProvider(id, a.dir, "m", "", "title", ""); err != nil {
		t.Fatal(err)
	}
	call := provider.ToolCall{ID: "c1", Type: "function"}
	call.Function.Name, call.Function.Arguments = "bash", "{}"
	calls := []provider.ToolCall{call}
	for _, msg := range []provider.Message{{Role: "user", Content: "go"}, {Role: "assistant", ToolCalls: calls}} {
		if err := a.store.AppendMessage(id, msg); err != nil {
			t.Fatal(err)
		}
	}
	rec, _, _ := a.store.GetSession(id)
	return rec
}

func TestOpeningASessionOfALiveProcessIsReadOnly(t *testing.T) {
	a := startingApp(t)
	rec := sessionWithOpenToolCall(t, a, "s1")
	claimAs(t, os.Getppid(), "s1")
	s, err := a.openSession(rec)
	if err != nil {
		t.Fatal(err)
	}
	if s.readOnlyPID != os.Getppid() {
		t.Fatalf("readOnlyPID = %d", s.readOnlyPID)
	}
	if got := s.history[len(s.history)-1].text; got != "Read-only: in use by process "+strconv.Itoa(os.Getppid()) {
		t.Errorf("message = %q", got)
	}
	msgs, _ := a.store.LoadMessages("s1")
	if len(msgs) != 2 {
		t.Errorf("fake interrupted results were appended: %d messages", len(msgs))
	}
	s.ready = true
	s.pending = append(s.pending, core.Request{Prompt: "hello"})
	a.flushPending()
	if len(s.pending) != 1 {
		t.Error("a read-only session handed a request to its agent")
	}
}

func TestOpeningASessionOfADeadProcessKeepsRecovering(t *testing.T) {
	a := startingApp(t)
	rec := sessionWithOpenToolCall(t, a, "s1")
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	claimAs(t, cmd.Process.Pid, "s1")
	s, err := a.openSession(rec)
	if err != nil {
		t.Fatal(err)
	}
	if s.readOnlyPID != 0 {
		t.Errorf("readOnlyPID = %d", s.readOnlyPID)
	}
	if msgs, _ := a.store.LoadMessages("s1"); len(msgs) != 3 {
		t.Errorf("messages = %d, want the interrupted result appended", len(msgs))
	}
}
