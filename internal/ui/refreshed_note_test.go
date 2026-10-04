package ui

import (
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

func TestResumeAndRewindHideTheRefreshedNote(t *testing.T) {
	content := "<system-refreshed>instructions were refreshed</system-refreshed>\n\n" +
		core.TodoEditedBlock([]todo.Item{{Text: "a", Status: todo.Done}}) + "<prompts>\nx\n</prompts>\n\nhi"
	msgs := []provider.Message{{Role: "user", Content: content}}
	entries := historyToEntries(msgs, tools.Build(tools.Catalog(), &tools.MemoryTodos{}))
	if len(entries) != 1 || entries[0].text != "hi" {
		t.Errorf("entries = %+v", entries)
	}
	if text, ok := typedText(msgs[0]); !ok || text != "hi" {
		t.Errorf("rewind text = %q, %v", text, ok)
	}
}
