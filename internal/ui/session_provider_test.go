package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func modelServer(t *testing.T, hits *int, models ...string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*hits++
		data := ""
		for i, m := range models {
			if i > 0 {
				data += ","
			}
			data += fmt.Sprintf(`{"id":%q}`, m)
		}
		fmt.Fprintf(w, `{"data":[%s]}`, data)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func twoProviderApp(t *testing.T, urlA, urlB string) *app {
	a := startingApp(t)
	a.cfg.Providers = []store.ProviderEntry{
		{ID: "a", Name: "a", BaseURL: urlA, APIKey: "k"},
		{ID: "b", Name: "b", BaseURL: urlB, APIKey: "k"},
	}
	if err := a.store.SaveProviders(a.cfg.Providers, "a"); err != nil {
		t.Fatal(err)
	}
	a.cfg.ActiveProvider = "a"
	a.cfg.Provider = provider.Config{BaseURL: urlA, APIKey: "k"}
	a.modelsLoaded = make(chan modelsResult, 1)
	return a
}

func TestClientForSeparatesEmptyKnownAndMissingProviders(t *testing.T) {
	a := twoProviderApp(t, "http://a.local", "http://b.local")
	cases := []struct {
		id, wantURL, wantID string
		missing             bool
	}{
		{"", "http://a.local", "a", false},
		{"b", "http://b.local", "b", false},
		{"gone", "", "gone", true},
	}
	for _, c := range cases {
		client, id, missing := a.clientFor(c.id)
		if client.Config().BaseURL != c.wantURL || id != c.wantID || missing != c.missing {
			t.Errorf("clientFor(%q) = %q, %q, %v", c.id, client.Config().BaseURL, id, missing)
		}
	}
}

func TestSessionOfDeletedProviderCannotSendAndRebindsOnPick(t *testing.T) {
	a := twoProviderApp(t, "http://a.local", "http://b.local")
	s := a.startSession("s1", "b", "m", "", nil, nil)
	a.active = s
	if !s.sessionReady() {
		t.Fatal("a session of a saved provider must be ready")
	}
	a.cfg.Providers = a.cfg.Providers[:1]
	a.markMissingProviders()
	if !s.providerMissing || s.sessionReady() || s.client.Config().Ready() {
		t.Fatalf("the session of a deleted provider must be cut off: %+v", s.client.Config())
	}
	if last := s.history[len(s.history)-1]; last.text == "" {
		t.Error("no message about the deleted provider")
	}
	if err := a.activateProvider("a"); err != nil {
		t.Fatal(err)
	}
	if s.providerMissing || s.provider != "a" || s.client.Config().BaseURL != "http://a.local" {
		t.Errorf("session after the pick: %q missing=%v", s.provider, s.providerMissing)
	}
}

func TestEmptyProviderSessionFollowsTheActiveProvider(t *testing.T) {
	a := twoProviderApp(t, "http://a.local", "http://b.local")
	s := a.startSession("s1", "", "m", "", nil, nil)
	if s.providerMissing || s.provider != "a" || s.client.Config().BaseURL != "http://a.local" {
		t.Errorf("empty provider id: %q missing=%v", s.provider, s.providerMissing)
	}
}
