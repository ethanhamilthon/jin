package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
	"jin/internal/provider"
)

func TestTellUpdatesShowMessageAndKeepSuggestion(t *testing.T) {
	s := &chatSession{width: 80, ready: true}
	s.showUpdate(core.Update{Kind: core.UpdateToolCall, Tool: "tell_user", Text: "message: hi"})
	s.showUpdate(core.Update{Kind: core.UpdateTell, Text: "Tests take two minutes"})
	s.showUpdate(core.Update{Kind: core.UpdateSuggest, Text: "Deploy it"})
	if len(s.history) != 1 || s.history[0].kind != core.UpdateTell {
		t.Fatalf("history = %+v", s.history)
	}
	if !s.showsSuggestion() {
		t.Fatal("an idle empty input must show the suggestion")
	}
	s.input = []string{"x"}
	if s.showsSuggestion() {
		t.Fatal("typing hides the suggestion")
	}
}

func TestRightPutsTheSuggestionIntoTheInput(t *testing.T) {
	a := &app{}
	s := &chatSession{ready: true, suggestion: "Deploy it"}
	a.active = s
	if !a.suggestionKey(tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)) {
		t.Fatal("→ was not handled")
	}
	if got := len(s.input); got == 0 || s.suggestion != "" {
		t.Fatalf("input %v suggestion %q", s.input, s.suggestion)
	}
	if a.suggestionKey(tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)) {
		t.Fatal("→ must move the cursor once the input has text")
	}
}

func TestSavedTellCallsShowAgain(t *testing.T) {
	call := func(id, args string) provider.ToolCall {
		c := provider.ToolCall{ID: id, Type: "function"}
		c.Function.Name, c.Function.Arguments = "tell_user", args
		return c
	}
	entries := historyToEntries([]provider.Message{{Role: "assistant", ToolCalls: []provider.ToolCall{
		call("1", `{"mode":"message","text":"Halfway"}`), call("2", `{"mode":"suggest","text":"Next"}`)}}}, nil)
	if len(entries) != 1 || entries[0].kind != core.UpdateTell || entries[0].text != "Halfway" {
		t.Fatalf("entries = %+v", entries)
	}
}
