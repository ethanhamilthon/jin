package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/pricing"
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
	a.showContext(false)
	last := s.history[len(s.history)-1]
	if last.kind != core.UpdateInfo || !strings.Contains(last.text, "Conversation: 1 messages") || !strings.Contains(last.text, "system prompt") {
		t.Errorf("entry = %q", last.text)
	}
}

func TestContextCommandCountsPrunedResultsAsOmitted(t *testing.T) {
	db, _ := openFoldDB(t)
	a, _ := layoutApp(t)
	a.dir = t.TempDir()
	s := a.active
	s.id, s.store, s.model = "pruned", db, "m"
	s.agent = core.NewAgent(nil, "You are jin.", tools.NewRegistry())
	s.pricing = pricing.Table{"m": {MaxInputTokens: 1200}}
	call := provider.ToolCall{ID: "1"}
	call.Function.Name = "read"
	db.AppendMessage("pruned", provider.Message{Role: "user", Content: "go"})
	db.AppendMessage("pruned", provider.Message{Role: "assistant", ToolCalls: []provider.ToolCall{call}})
	db.AppendMessage("pruned", provider.Message{Role: "tool", ToolCallID: "1", Content: strings.Repeat("x", 4000)})
	for range 4 {
		db.AppendMessage("pruned", provider.Message{Role: "user", Content: "more"})
	}
	a.showContext(false)
	text := s.history[len(s.history)-1].text
	if !strings.Contains(text, "read") || strings.Contains(text, "~1K") {
		t.Errorf("pruned result still counted in full:\n%s", text)
	}
}

func TestContextFullPrintsTheTextOfEveryPart(t *testing.T) {
	db, _ := openFoldDB(t)
	a, _ := layoutApp(t)
	a.dir = t.TempDir()
	s := a.active
	s.id, s.store = "ctx", db
	prompt := core.BuildSystemPrompt([]core.PromptPart{{Name: "system text", Text: "You are jin.\n\n"}, {Name: "command: jin docs", Text: "Docs pointer."}})
	s.agent = core.NewAgent(nil, prompt, tools.NewRegistry(tools.NewAsk()))
	a.showContext(true)
	last := s.history[len(s.history)-1].text
	for _, want := range []string{"--- system text ---\nYou are jin.", "--- command: jin docs ---\nDocs pointer.", "--- tool schema: ask_user ---\n{"} {
		if !strings.Contains(last, want) {
			t.Errorf("full context lacks %q:\n%s", want, last)
		}
	}
	a.showContext(false)
	if strings.Contains(s.history[len(s.history)-1].text, "Docs pointer.") {
		t.Error("the short form must not print the texts")
	}
}
