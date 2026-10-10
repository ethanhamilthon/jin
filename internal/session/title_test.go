package session

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"jin/internal/core"
	"jin/internal/store"
)

// titleMarker is in the title prompt, so the fake provider can tell title
// requests from chat requests.
const titleMarker = "NAME-THIS-CHAT"

func reply(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":%q}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n", text)
}

// titleServer answers title requests with "Greeting chat." (or fails with
// 400, which the provider does not retry) and counts them.
func titleServer(calls *atomic.Int32, fail bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), titleMarker) {
			reply(w, "hello there")
			return
		}
		calls.Add(1)
		if fail {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"no"}}`))
			return
		}
		reply(w, "Greeting chat.")
	}
}

func TestGenerateTitleNamesTheSession(t *testing.T) {
	var calls atomic.Int32
	m, events, dir := newManagerWith(t, titleServer(&calls, false))
	if err := m.DB().SaveTitle(store.TitleSettings{Prompt: titleMarker, After: 0}); err != nil {
		t.Fatal(err)
	}
	snap, _ := m.Create(dir)
	id := snap.State.ID
	waitFor(t, events, func(ev Event) bool { return ev.Type == "state" && ev.State.Ready })
	if _, err := m.GenerateTitle(context.Background(), id); err == nil {
		t.Fatal("an empty session was titled")
	}
	if err := m.Send(id, "say hi", nil, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "ring" && ev.Kind == core.UpdateDone })
	title, err := m.GenerateTitle(context.Background(), id)
	if err != nil || title != "Greeting chat" {
		t.Fatalf("title = %q, %v", title, err)
	}
	if snap, _ := m.Snapshot(id); snap.State.Title != "Greeting chat" {
		t.Fatalf("live title = %q", snap.State.Title)
	}
	if rec, _, _ := m.DB().GetSession(id); rec.Title != "Greeting chat" {
		t.Fatalf("stored title = %q", rec.Title)
	}
	if calls.Load() != 1 {
		t.Fatalf("title calls = %d", calls.Load())
	}
}

func TestAutoTitleRunsOnceAtTheCount(t *testing.T) {
	var calls atomic.Int32
	m, events, dir := newManagerWith(t, titleServer(&calls, false))
	if err := m.DB().SaveTitle(store.TitleSettings{Prompt: titleMarker, After: 1}); err != nil {
		t.Fatal(err)
	}
	snap, _ := m.Create(dir)
	id := snap.State.ID
	waitFor(t, events, func(ev Event) bool { return ev.Type == "state" && ev.State.Ready })
	if err := m.Send(id, "say hi", nil, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "ring" && ev.Kind == core.UpdateDone })
	waitFor(t, events, func(ev Event) bool { return ev.Type == "state" && ev.State.Title == "Greeting chat" })
	if err := m.Send(id, "again", nil, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "ring" && ev.Kind == core.UpdateDone })
	if calls.Load() != 1 {
		t.Fatalf("title calls = %d, want 1", calls.Load())
	}
}

func TestAutoTitleFailureIsReportedWithoutRetry(t *testing.T) {
	var calls atomic.Int32
	m, events, dir := newManagerWith(t, titleServer(&calls, true))
	if err := m.DB().SaveTitle(store.TitleSettings{Prompt: titleMarker, After: 1}); err != nil {
		t.Fatal(err)
	}
	snap, _ := m.Create(dir)
	id := snap.State.ID
	waitFor(t, events, func(ev Event) bool { return ev.Type == "state" && ev.State.Ready })
	if err := m.Send(id, "say hi", nil, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool {
		return ev.Type == "notice" && strings.HasPrefix(ev.Text, "Title was not generated")
	})
	if calls.Load() != 1 {
		t.Fatalf("title calls = %d, want 1", calls.Load())
	}
	if snap, _ := m.Snapshot(id); snap.State.Title != "say hi" {
		t.Fatalf("title = %q", snap.State.Title)
	}
}

func TestCleanTitle(t *testing.T) {
	cases := map[string]string{
		"  \"Parser bug.\"  \n\nmore text": "Parser bug",
		"'Go modules'":                     "Go modules",
		"**Release notes**":                "Release notes",
	}
	for in, want := range cases {
		if got := cleanTitle(in); got != want {
			t.Errorf("cleanTitle(%q) = %q, want %q", in, got, want)
		}
	}
	if got := cleanTitle(strings.Repeat("a", 80)); len([]rune(got)) != maxTitle {
		t.Errorf("long title has %d runes", len([]rune(got)))
	}
}
