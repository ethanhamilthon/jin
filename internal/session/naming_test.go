package session

import "testing"

func TestCleanTitleDoesNotIncludeInstruction(t *testing.T) {
	if got := cleanTitle("  \"Parser bug.\"  "); got != "Parser bug" {
		t.Fatalf("title = %q", got)
	}
}
