package ui

import (
	"testing"
	"time"

	"jin/internal/provider"
)

func pumpLoad(t *testing.T, a *app) {
	t.Helper()
	select {
	case result := <-a.loads:
		a.receiveLoad(result)
	case <-time.After(5 * time.Second):
		t.Fatal("no picker result")
	}
}

func pickerApp(t *testing.T) (*app, *chatSession, *int, *int) {
	var hitsA, hitsB int
	a := twoProviderApp(t, modelServer(t, &hitsA, "a1", "a2"), modelServer(t, &hitsB, "b1", "b2"))
	a.loads = make(chan loadResult, 4)
	a.cfg.Model = "a1"
	s := &chatSession{id: "s1", provider: "b", model: "b1", client: provider.NewClient(provider.Config{BaseURL: a.cfg.Providers[1].BaseURL, APIKey: "k"})}
	a.active = s
	return a, s, &hitsA, &hitsB
}

func choose(t *testing.T, a *app, model string) {
	t.Helper()
	pumpLoad(t, a)
	if err := a.sel.submit(model); err != nil {
		t.Fatal(err)
	}
	pumpLoad(t, a)
}

func TestModelFlowUsesTheSessionProviderAndKeepsTheGlobalModel(t *testing.T) {
	a, s, hitsA, hitsB := pickerApp(t)
	a.openModelFlow()
	choose(t, a, "b2")
	if err := a.sel.submit(defaultEffort); err != nil {
		t.Fatal(err)
	}
	if *hitsA != 0 || *hitsB == 0 {
		t.Errorf("requests: a=%d b=%d", *hitsA, *hitsB)
	}
	if s.model != "b2" || a.cfg.Model != "a1" {
		t.Errorf("session model %q, global model %q", s.model, a.cfg.Model)
	}
}

func TestModelPickerResultAfterASessionSwitchIsIgnored(t *testing.T) {
	a, s, _, _ := pickerApp(t)
	a.openModelFlow()
	choose(t, a, "b2")
	other := &chatSession{id: "s2", provider: "a", model: "a1"}
	a.active = other
	if err := a.sel.submit(defaultEffort); err != nil {
		t.Fatal(err)
	}
	if s.model != "b1" || other.model != "a1" || a.cfg.Model != "a1" {
		t.Errorf("models: %q %q global %q", s.model, other.model, a.cfg.Model)
	}
}
