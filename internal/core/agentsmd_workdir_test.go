package core

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNestedAgentsUseEachAgentsWorkdir(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	for dir, rule := range map[string]string{first: "first rules", second: "second rules"} {
		writeFile(t, filepath.Join(dir, "child", "AGENTS.md"), rule)
		writeFile(t, filepath.Join(dir, "child", "file.txt"), "content")
	}
	firstAgent, secondAgent := NewAgent(nil, "sys", nil), NewAgent(nil, "sys", nil)
	firstAgent.SetWorkdir(first)
	secondAgent.SetWorkdir(second)
	call := readCall(t, "1", "child/file.txt")
	firstText := firstAgent.nestedAgents(call, "ok")
	secondText := secondAgent.nestedAgents(call, "ok")
	if !strings.Contains(firstText, "first rules") || strings.Contains(firstText, "second rules") {
		t.Fatalf("first session instructions = %q", firstText)
	}
	if !strings.Contains(secondText, "second rules") || strings.Contains(secondText, "first rules") {
		t.Fatalf("second session instructions = %q", secondText)
	}
}
