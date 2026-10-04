package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tools"
)

func TestContextInfoText(t *testing.T) {
	info := contextInfo{
		used: 20000, window: 100000,
		prompt: []core.PromptPart{
			{Name: "system text", Text: strings.Repeat("a", 4000)},
			{Name: "AGENTS.md /w/AGENTS.md", Text: strings.Repeat("b", 400)},
			{Name: "cache break", Text: provider.CacheBreak},
		},
		toolSchemas: 8000, messages: 3, conversation: 2000,
		results: []contextResult{{"read main.go", 4000}},
		cache:   cacheRate{percent: 80, known: true},
	}
	text := info.String()
	for _, want := range []string{
		"approximate", "Used: 20K of 100K (20%)", "System prompt: ~1.1K", "  system text: ~1K",
		"  AGENTS.md /w/AGENTS.md: ~100", "Tool schemas: ~2K", "Conversation: 3 messages, ~500",
		"Largest tool results:\n  read main.go: ~1K", "Cache: 80% of the last request",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "cache break") {
		t.Error("the cache break marker is not part of the request")
	}
}

func TestContextInfoWithoutOptionalLines(t *testing.T) {
	text := contextInfo{used: 500}.String()
	if strings.Contains(text, "Largest") || strings.Contains(text, "Cache:") || strings.Contains(text, "%") {
		t.Errorf("unexpected lines:\n%s", text)
	}
}

func TestContextCommandPrintsTheBlock(t *testing.T) {
	db, _ := openFoldDB(t)
	a, _ := layoutApp(t)
	a.dir = t.TempDir()
	s := a.active
	s.id, s.store = "ctx", db
	s.agent = core.NewAgent(nil, "You are jin.\n\n"+provider.CacheBreak+"\n\nEnvironment:", tools.NewRegistry())
	s.usage.Context = 1000
	db.AppendMessage("ctx", provider.Message{Role: "user", Content: "hello"})
	a.showContext()
	last := s.history[len(s.history)-1]
	if last.kind != core.UpdateInfo || !strings.Contains(last.text, "Conversation: 1 messages") || !strings.Contains(last.text, "system text") {
		t.Errorf("entry = %q", last.text)
	}
}
