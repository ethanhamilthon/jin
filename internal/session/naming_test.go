package session

import "testing"

func TestCleanTitleDoesNotIncludeInstruction(t *testing.T) {
	if got := cleanTitle("  \"Parser bug.\"  "); got != "Parser bug" {
		t.Fatalf("title = %q", got)
	}
}

func TestPlausibleTitleRefusesReplies(t *testing.T) {
	long := "Привет! Я Claude Code, CLI-помощник для работы с кодом. Я помогаю читать файлы и запускать команды в вашем проекте"
	for answer, want := range map[string]bool{
		"Parser bug": true,
		"I don't have a view of your working directory yet. I can list files": false,
		"Title\nExtra explanation": false,
		long:                       false,
	} {
		if got := plausibleTitle(answer, cleanTitle(answer)); got != want {
			t.Errorf("%q = %v, want %v", answer, got, want)
		}
	}
}
