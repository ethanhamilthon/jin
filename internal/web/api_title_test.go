package web

import (
	"strings"
	"testing"
)

func TestTitleRoutesAreUnavailableWithoutSideEffects(t *testing.T) {
	ts, _ := newTestServer(t)
	var failure map[string]string
	for _, path := range []string{"/api/settings/title", "/api/sessions/nope/generate-title"} {
		if code := call(t, ts, "POST", path, `{"prompt":"changed"}`, &failure); code != 400 || !strings.Contains(failure["error"], "unavailable") {
			t.Fatalf("%s: %d %v", path, code, failure)
		}
	}
	var state map[string]any
	if code := call(t, ts, "GET", "/api/state", "", &state); code != 200 {
		t.Fatalf("state: %d", code)
	}
	config := state["config"].(map[string]any)
	if _, ok := config["title"]; ok {
		t.Fatal("state exposes title configuration")
	}
}
