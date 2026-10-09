package ui

import (
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func TestResumeAndRewindHideTheRefreshedNote(t *testing.T) {
	content := "<system-refreshed>instructions were refreshed</system-refreshed>\n\n" +
		"<todo-edited>x</todo-edited>\n\n" + "<pasted-prompts>\nx\n</pasted-prompts>\n\nhi"
	msgs := []provider.Message{{Role: "user", Content: content}}
	entries := historyToEntries(msgs, tools.Build(tools.Catalog()))
	if len(entries) != 1 || entries[0].text != "hi" {
		t.Errorf("entries = %+v", entries)
	}
	if text, ok := typedText(msgs[0]); !ok || text != "hi" {
		t.Errorf("rewind text = %q, %v", text, ok)
	}
}
