package headless

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"jin/internal/store"
)

// sessionHarness has a saved session of two turns and the provider saved, as
// jin sessions <action> reads it from the data folder.
func sessionHarness(t *testing.T) (*harness, string) {
	t.Helper()
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	entry := store.ProviderEntry{ID: "p", Name: "p", Kind: "openai", BaseURL: h.env["JIN_BASE_URL"], APIKey: "k"}
	if err := h.db.AddProvider(entry); err != nil {
		t.Fatal(err)
	}
	if err := h.db.SetActiveProvider("p"); err != nil {
		t.Fatal(err)
	}
	if err := h.db.SaveModel("m", ""); err != nil {
		t.Fatal(err)
	}
	h.dir = t.TempDir()
	for _, prompt := range [][]string{{"-p", "first"}, {"-p", "-c", "second"}} {
		if code := h.run(t, prompt...); code != 0 {
			t.Fatalf("setup run %v: code %d, stderr %q", prompt, code, h.errOut.String())
		}
	}
	list, _ := h.db.ListByPath(h.dir)
	if len(list) != 1 {
		t.Fatalf("sessions = %+v", list)
	}
	return h, list[0].ID
}

func TestSessionsContextTextAndJSON(t *testing.T) {
	h, id := sessionHarness(t)
	if code := h.run(t, "sessions", "context", id[:8]); code != 0 || !strings.HasPrefix(h.out.String(), "context: ") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, h.out.String(), h.errOut.String())
	}
	if code := h.run(t, "sessions", "context", id, "--format", "json"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	var report map[string]any
	if err := json.Unmarshal(h.out.Bytes(), &report); err != nil || report["window"] == nil {
		t.Fatalf("json = %q (%v)", h.out.String(), err)
	}
}

func TestSessionsRewindListsAndForks(t *testing.T) {
	h, id := sessionHarness(t)
	if code := h.run(t, "sessions", "rewind", id); code != 0 || h.out.String() != "1\tfirst\n2\tsecond\n" {
		t.Fatalf("code %d, stdout %q, stderr %q", code, h.out.String(), h.errOut.String())
	}
	if code := h.run(t, "sessions", "rewind", id, "--to", "2"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	next := strings.TrimSpace(h.out.String())
	messages, err := h.db.LoadMessages(next)
	if err != nil || len(messages) != 2 || next == id || !strings.Contains(h.errOut.String(), "message: second") {
		t.Fatalf("fork %q has %d messages (%v), stderr %q", next, len(messages), err, h.errOut.String())
	}
}

func TestSessionsCompactSummarizes(t *testing.T) {
	h, id := sessionHarness(t)
	before, _ := h.db.LoadMessages(id)
	if code := h.run(t, "sessions", "compact", id); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	after, _ := h.db.LoadMessages(id)
	if len(after) <= len(before) {
		t.Fatalf("messages before %d, after %d: no summary was saved", len(before), len(after))
	}
}

func TestSessionsHandoffPrintsTheBrief(t *testing.T) {
	h, id := sessionHarness(t)
	if code := h.run(t, "sessions", "handoff", id); code != 0 || h.out.String() != "hello there\n" {
		t.Fatalf("code %d, stdout %q, stderr %q", code, h.out.String(), h.errOut.String())
	}
}

func TestSessionsUndoWithoutChangesFails(t *testing.T) {
	h, id := sessionHarness(t)
	if code := h.run(t, "sessions", "undo", id); code != 1 || !strings.Contains(h.errOut.String(), "Nothing to undo") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, h.out.String(), h.errOut.String())
	}
}

func TestSessionsReloadReports(t *testing.T) {
	h, id := sessionHarness(t)
	if code := h.run(t, "sessions", "reload", id); code != 0 || !strings.Contains(h.out.String(), "Reloaded") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, h.out.String(), h.errOut.String())
	}
}

func TestSessionsActionErrors(t *testing.T) {
	h, _ := sessionHarness(t)
	for _, args := range [][]string{
		{"sessions", "undo", "nosuchid"},
		{"sessions", "undo"},
		{"sessions", "context", "x", "--format", "xml"},
		{"sessions", "undo", "x", "--to", "1"},
	} {
		if code := h.run(t, args...); code != 1 || h.errOut.Len() == 0 {
			t.Errorf("%v: code %d, stderr %q", args, code, h.errOut.String())
		}
	}
}
