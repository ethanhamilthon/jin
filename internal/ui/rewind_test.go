package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/files"
	"jin/internal/provider"
	"jin/internal/todo"
)

func TestRewindPointsSkipInjectedMessages(t *testing.T) {
	messages := []provider.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "first"},
		{Role: "assistant", Content: "ok"},
		{Role: "user", Content: "<task-result id=\"x\">done</task-result>"},
		{Role: "user", Content: core.UndoBlock([]string{"f"}) + core.TodoEditedBlock([]todo.Item{{Text: "a", Status: todo.Done}}) + "second\n\n" + files.Block([]string{"/a"})},
	}
	points := rewindPoints(messages)
	if len(points) != 2 || points[0].text != "first" || points[1].text != "second" || points[1].index != 4 {
		t.Fatalf("points = %+v", points)
	}
}

func TestForkAtSavesHistoryAndDraft(t *testing.T) {
	db, _ := openFoldDB(t)
	a, _ := layoutApp(t)
	a.store, a.ctx, a.dir = db, t.Context(), t.TempDir()
	a.updates, a.rendered = make(chan taggedUpdate, 8), make(chan renderEvent, 8)
	a.active.persisted = true
	from := &chatSession{id: "old", path: a.dir, model: "m", title: "t"}
	history := []provider.Message{{Role: "user", Content: "first"}, {Role: "assistant", Content: "ok"}}
	if err := a.forkAt(from, history, "second"); err != nil {
		t.Fatal(err)
	}
	if a.active.id == "old" || strings.Join(a.active.input, "") != "second" || !a.active.persisted {
		t.Fatalf("active = %+v", a.active)
	}
	saved, _ := db.LoadMessages(a.active.id)
	if len(saved) != 2 || saved[1].Content != "ok" {
		t.Fatalf("saved = %+v", saved)
	}
}
