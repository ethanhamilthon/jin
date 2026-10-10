package ui

import (
	"slices"
	"testing"
	"time"

	"jin/internal/provider"
)

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
	if hitsA != 1 || hitsB != 1 {
		t.Errorf("requests: a=%d b=%d", hitsA, hitsB)
	}
	if s.model != "b2" || a.cfg.Model != "a1" {
		t.Errorf("session model %q, global model %q", s.model, a.cfg.Model)
	}
	a.cycleModel()
	if hitsB != 1 || s.model != "a1" || s.provider != "a" {
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
