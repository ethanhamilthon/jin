package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

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

func TestCycleModelUsesTheSessionProviderAndKeepsTheGlobalModel(t *testing.T) {
	var hitsA, hitsB int
	a := twoProviderApp(t, modelServer(t, &hitsA, "a1", "a2"), modelServer(t, &hitsB, "b1", "b2"))
	a.cfg.Model = "a1"
	if err := a.store.SaveScopeFor("b", []string{"b1", "b2"}); err != nil {
		t.Fatal(err)
	}
	s := &chatSession{id: "s1", provider: "b", model: "b1", client: provider.NewClient(provider.Config{BaseURL: a.cfg.Providers[1].BaseURL, APIKey: "k"})}
	a.active = s
	a.cycleModel()
	select {
	case result := <-a.modelsLoaded:
		a.receiveModels(result)
	case <-time.After(5 * time.Second):
		t.Fatal("no model list")
	}
	if hitsA != 0 || hitsB != 1 {
		t.Errorf("requests: a=%d b=%d", hitsA, hitsB)
	}
	if s.model != "b2" || a.cfg.Model != "a1" {
		t.Errorf("session model %q, global model %q", s.model, a.cfg.Model)
	}
	a.cycleModel()
	if hitsB != 1 || s.model != "b1" {
		t.Errorf("the list must be cached per provider: hits=%d model=%q", hitsB, s.model)
	}
}

func TestStaleModelListAfterSessionSwitchIsIgnored(t *testing.T) {
	a := twoProviderApp(t, "http://a.local", "http://b.local")
	one := &chatSession{id: "one", provider: "a", model: "a1"}
	two := &chatSession{id: "two", provider: "a", model: "a1"}
	a.active, a.loadingModels = two, true
	a.receiveModels(modelsResult{session: "one", provider: "a", models: []string{"a1", "a2"}})
	if a.loadingModels || two.models != nil || two.model != "a1" || one.model != "a1" {
		t.Errorf("stale list applied: %+v", two)
	}
	a.active = &chatSession{id: "one", provider: "b", model: "b1"}
	a.receiveModels(modelsResult{session: "one", provider: "a", models: []string{"x"}})
	if a.active.models != nil {
		t.Error("a list of another provider was cached")
	}
}

func TestSessionScopeIsTheScopeOfItsOwnProvider(t *testing.T) {
	a := twoProviderApp(t, "http://a.local", "http://b.local")
	a.cfg.Scope = []string{"a1"}
	if err := a.store.SaveScopeFor("b", []string{"b2"}); err != nil {
		t.Fatal(err)
	}
	if got := a.sessionScope(&chatSession{provider: "b"}); !slices.Equal(got, []string{"b2"}) {
		t.Errorf("scope of b = %v", got)
	}
	if got := a.sessionScope(&chatSession{provider: "a"}); !slices.Equal(got, []string{"a1"}) {
		t.Errorf("scope of a = %v", got)
	}
}
