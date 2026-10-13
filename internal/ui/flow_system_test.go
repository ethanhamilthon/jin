package ui

import (
	"testing"
)

func TestSystemPromptSettingIsNotAvailable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a, _ := layoutApp(t)
	for _, entry := range a.settingsEntries() {
		if entry.label == "System prompt" {
			t.Fatal("system prompt editing must not appear in settings")
		}
	}

}
