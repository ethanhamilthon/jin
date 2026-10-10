package web

import (
	"strings"
	"testing"

	"jin/internal/store"
)

func TestTitleSettingsRoundTrip(t *testing.T) {
	ts, _ := newTestServer(t)
	var failure map[string]string
	if code := call(t, ts, "POST", "/api/settings/title", `{"model":"m","after":3}`, &failure); code != 400 || !strings.Contains(failure["error"], "provider") {
		t.Fatalf("model without provider: %d %v", code, failure)
	}
	if code := call(t, ts, "POST", "/api/settings/title", `{"after":51}`, &failure); code != 400 {
		t.Fatalf("after 51: %d", code)
	}
	if code := call(t, ts, "POST", "/api/settings/title", `{"after":3,"prompt":"Name it."}`, &failure); code != 200 {
		t.Fatalf("save: %d %v", code, failure)
	}
	var state struct{ Config configView }
	if code := call(t, ts, "GET", "/api/state", "", &state); code != 200 || state.Config.Title.After != 3 || state.Config.Title.Prompt != "Name it." {
		t.Fatalf("state: %d %+v", code, state.Config.Title)
	}
}

func TestDefaultTitleSettingsInState(t *testing.T) {
	ts, _ := newTestServer(t)
	var state struct{ Config configView }
	call(t, ts, "GET", "/api/state", "", &state)
	if state.Config.Title.After != store.DefaultTitleAfter || state.Config.Title.Prompt != store.DefaultTitlePrompt {
		t.Fatalf("defaults = %+v", state.Config.Title)
	}
}

func TestGenerateTitleNeedsAnOpenSession(t *testing.T) {
	ts, _ := newTestServer(t)
	var failure map[string]string
	if code := call(t, ts, "POST", "/api/sessions/nope/generate-title", "", &failure); code != 400 || !strings.Contains(failure["error"], "not open") {
		t.Fatalf("generate on unknown session: %d %v", code, failure)
	}
}
