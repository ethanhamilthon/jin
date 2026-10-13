package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/session"
	"jin/internal/store"
	"jin/internal/sysprompt"
)

func newTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := &server{ctx: ctx, db: db, hub: newHub(), dir: dir, version: "v0", quit: cancel}
	s.m = session.NewManager(ctx, db, "v0", func(ev session.Event) { s.publish(ev) })
	t.Cleanup(func() { cancel(); s.m.Shutdown() })
	ts := httptest.NewServer(s.routes())
	t.Cleanup(ts.Close)
	return ts, dir
}

func call(t *testing.T, ts *httptest.Server, method, path, body string, out any) int {
	t.Helper()
	req, _ := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func TestStateAndNewSession(t *testing.T) {
	ts, dir := newTestServer(t)
	var state struct {
		Version string
		Dir     string
		Config  configView
	}
	if code := call(t, ts, "GET", "/api/state", "", &state); code != 200 || state.Dir != dir || state.Config.Ready || len(state.Config.Tools) == 0 {
		t.Fatalf("state %d %+v", code, state)
	}
	var snap session.Snapshot
	if code := call(t, ts, "POST", "/api/sessions", `{"path":"`+dir+`"}`, &snap); code != 200 || snap.State.Path != dir || snap.Intro == nil {
		t.Fatalf("new session %d %+v", code, snap)
	}
	var failure map[string]string
	if code := call(t, ts, "POST", "/api/sessions/nope/send", `{"text":"hi"}`, &failure); code != 400 || failure["error"] == "" {
		t.Fatalf("send to unknown %d %v", code, failure)
	}
}

func TestPromptsRoundTrip(t *testing.T) {
	ts, _ := newTestServer(t)
	var failure map[string]string
	if code := call(t, ts, "POST", "/api/prompts", `{"name":"notes","content":"Review the diff"}`, &failure); code != 200 {
		t.Fatalf("create %d", code)
	}
	var text textView
	if call(t, ts, "GET", "/api/prompts/text?name=notes", "", &text); text.Content != "Review the diff" {
		t.Fatalf("text %+v", text)
	}
	var list []promptView
	call(t, ts, "GET", "/api/prompts", "", &list)
	found := false
	for _, p := range list {
		found = found || p.Name == "notes" && p.Enabled
	}
	if !found {
		t.Fatalf("list %+v", list)
	}
}

func TestSysPromptTemplatesAreReadOnlyAndMutationsAreUnavailable(t *testing.T) {
	ts, _ := newTestServer(t)
	path, err := sysprompt.Path()
	if err != nil {
		t.Fatal(err)
	}
	var view textView
	if code := call(t, ts, "GET", "/api/sysprompt", "", &view); code != http.StatusOK || view.Path != "" || view.Content == "" {
		t.Fatalf("templates: %d %+v", code, view)
	}
	var failure map[string]string
	if code := call(t, ts, "POST", "/api/sysprompt", `{"content":"changed"}`, &failure); code != http.StatusBadRequest || !strings.Contains(failure["error"], "unavailable") {
		t.Fatalf("save: %d %v", code, failure)
	}
	if code := call(t, ts, "POST", "/api/sysprompt/reset", `{"section":"all"}`, &failure); code != http.StatusBadRequest || !strings.Contains(failure["error"], "unavailable") {
		t.Fatalf("reset: %d %v", code, failure)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("legacy prompt file created or changed: %v", err)
	}
}

func TestArchiveHidesAndRestoresAProject(t *testing.T) {
	ts, _ := newTestServer(t)
	other := t.TempDir()
	if code := call(t, ts, "POST", "/api/projects", `{"path":"`+other+`"}`, nil); code != 200 {
		t.Fatalf("add: %d", code)
	}
	archived := func() bool {
		var list []struct {
			Path     string
			Archived bool
		}
		call(t, ts, "GET", "/api/projects", "", &list)
		for _, p := range list {
			if strings.HasSuffix(p.Path, filepath.Base(other)) {
				return p.Archived
			}
		}
		t.Fatal("project not listed")
		return false
	}
	if archived() {
		t.Fatal("a new project is archived")
	}
	call(t, ts, "POST", "/api/projects/archive", `{"path":"`+other+`","archived":true}`, nil)
	if !archived() {
		t.Fatal("archive did not stick")
	}
	call(t, ts, "POST", "/api/projects/archive", `{"path":"`+other+`","archived":false}`, nil)
	if archived() {
		t.Fatal("restore did not stick")
	}
}
