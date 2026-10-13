package ui

import "testing"

func TestSessionTitlesAreNotAvailableInSettings(t *testing.T) {
	a, _ := layoutApp(t)
	for _, entry := range a.settingsEntries() {
		if entry.label == "Session titles" {
			t.Fatal("session title settings must not appear")
		}
	}
}
