package prompts

import "testing"

func TestSummary(t *testing.T) {
	got := Summary("---\n# Terse mode\n\n> Respond  terse.\n- keep facts\n")
	if got != "Terse mode Respond terse. keep facts" {
		t.Fatalf("summary = %q", got)
	}
	if Preview("plan") == "" {
		t.Fatal("system prompt preview is empty")
	}
}
