package headless

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jin/internal/store"
)

func TestResumeUsesTheSessionProvider(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {})
	hits := map[string]int{}
	for _, id := range []string{"a", "b"} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits[id]++
			sse(w, answerChunk)
		}))
		t.Cleanup(server.Close)
		if err := h.db.AddProvider(store.ProviderEntry{ID: id, Name: id, BaseURL: server.URL, APIKey: "k"}); err != nil {
			t.Fatal(err)
		}
	}
	h.env = map[string]string{"JIN_MODEL": "m"}
	if err := h.db.SetActiveProvider("a"); err != nil {
		t.Fatal(err)
	}
	if code := h.run(t, "-p", "first"); code != 0 {
		t.Fatalf("first: %d %s", code, h.errOut.String())
	}
	list, _ := h.db.ListByPath(h.dir)
	if list[0].Provider != "a" {
		t.Fatalf("pinned provider %q", list[0].Provider)
	}
	_ = h.db.SetActiveProvider("b")
	if code := h.run(t, "-p", "-c", "second"); code != 0 {
		t.Fatalf("resume: %d %s", code, h.errOut.String())
	}
	if hits["a"] != 2 || hits["b"] != 0 {
		t.Fatalf("hits %v", hits)
	}
}

func TestResumeFailsClosedWhenProviderIsGone(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	_ = h.db.AddProvider(store.ProviderEntry{ID: "b", Name: "b", BaseURL: h.env["JIN_BASE_URL"], APIKey: "k"})
	_ = h.db.SetActiveProvider("b")
	if err := h.db.TouchProvider("s1", h.dir, "m", "", "t", "gone"); err != nil {
		t.Fatal(err)
	}
	env := h.env
	h.env = map[string]string{"JIN_MODEL": "m"}
	want := `session provider "gone" no longer exists; set JIN_BASE_URL and JIN_API_KEY to override`
	if code := h.run(t, "-p", "--session", "s1", "x"); code != exitError || !strings.Contains(h.errOut.String(), want) {
		t.Fatalf("code %d stderr %q", code, h.errOut.String())
	}
	h.env = env
	if code := h.run(t, "-p", "--session", "s1", "x"); code != 0 {
		t.Fatalf("override: %d %s", code, h.errOut.String())
	}
}
