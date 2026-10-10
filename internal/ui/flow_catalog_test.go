package ui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/provider"
	"jin/internal/sources"
)

func TestCatalogPickerChangesTheModelOfTheSession(t *testing.T) {
	var hitsA, hitsB int
	a := twoProviderApp(t, modelServer(t, &hitsA, "a1", "a2"), modelServer(t, &hitsB, "b1", "b2"))
	a.loads = make(chan loadResult, 4)
	a.cfg.Model = "a1"
	s := &chatSession{id: "s1", provider: "a", model: "a1", client: provider.NewClient(a.cfg.Provider)}
	a.active = s
	a.openModelFlow()
	step := func(what string) {
		t.Helper()
		select {
		case result := <-a.loads:
			a.receiveLoad(result)
		case <-time.After(5 * time.Second):
			t.Fatalf("%s did not load", what)
		}
	}
	step("models")
	a.sel.selectValue(sources.Key("b", "b2"))
	a.handleEvent(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	step("efforts")
	a.handleEvent(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if s.provider != "b" || s.model != "b2" {
		t.Fatalf("session = %s/%s, picker state %+v", s.provider, s.model, a.sel)
	}
}
