package ui

import (
	"errors"
	"testing"

	"jin/internal/core"
	"jin/internal/session"
	"jin/internal/store"
	"jin/internal/tools"
)

func TestBackendStreamingDoesNotDuplicateEntries(t *testing.T) {
	s := &chatSession{id: "shared", width: 80, ready: true, persisted: true}
	a := &app{sessions: map[string]*chatSession{s.id: s}}
	a.receiveBackend(session.Event{Type: "entry", Session: s.id, Seq: 1, Entry: &session.Entry{Kind: core.UpdateAssistant}})
	a.receiveBackend(session.Event{Type: "delta", Session: s.id, Seq: 2, Index: 0, Text: "hello"})
	a.receiveBackend(session.Event{Type: "delta", Session: s.id, Seq: 2, Index: 0, Text: "duplicate"})
	a.receiveBackend(session.Event{Type: "delta", Session: s.id, Seq: 3, Index: 0, Text: " world"})
	if len(s.history) != 1 || s.history[0].text != "hello world" {
		t.Fatalf("history: %+v", s.history)
	}
	if s.remoteEntries[0].Text != "hello world" {
		t.Fatalf("remote entries: %+v", s.remoteEntries)
	}
}

// A row this client adds itself, such as an error, must not receive the text
// that streams into an entry of the daemon.
func TestBackendDeltaFollowsItsEntryNotTheLastRow(t *testing.T) {
	s := &chatSession{id: "shared", width: 80, ready: true, persisted: true}
	a := &app{sessions: map[string]*chatSession{s.id: s}, cfg: store.Config{}}
	a.active = s
	a.receiveBackend(session.Event{Type: "entry", Session: s.id, Seq: 1, Entry: &session.Entry{Kind: core.UpdateAssistant}})
	a.backendError(errors.New("local note"))
	a.receiveBackend(session.Event{Type: "delta", Session: s.id, Seq: 2, Index: 0, Text: "hello"})
	if len(s.history) != 2 {
		t.Fatalf("history: %+v", s.history)
	}
	if s.history[0].text != "hello" || s.history[1].text != "local note" {
		t.Fatalf("delta landed on the wrong row: %+v", s.history)
	}
}

func TestBackendQuestionPreservesLocalAnswerDraft(t *testing.T) {
	s := &chatSession{id: "shared"}
	a := &app{}
	state := session.State{ID: s.id, Question: 1, Ask: []tools.Question{{Question: "choose"}}}
	a.backendState(s, state)
	s.ask.text = clusters("my draft")
	a.backendState(s, state)
	if len(s.ask.text) == 0 {
		t.Fatal("state update erased answer draft")
	}
	state.Question = 2
	a.backendState(s, state)
	if len(s.ask.text) != 0 {
		t.Fatal("new question retained old draft")
	}
	state.Ask = nil
	a.backendState(s, state)
	if s.ask != nil {
		t.Fatal("answered question remains open")
	}
}
