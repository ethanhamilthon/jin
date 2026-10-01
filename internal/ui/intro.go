package ui

import (
	"strings"

	"jin/internal/core"
)

const keysHelp = "INSERT  Esc normal · Enter send · Shift+Enter newline · Ctrl+C interrupt\n" +
	"NORMAL  Space commands · i insert · j/k scroll · m model · p provider · w search · s sessions · n new · x interrupt · q quit"

func (a *app) introEntries() []chatEntry {
	entries := []chatEntry{
		section("Tools", strings.Join(a.registry.Names(), ", ")),
		section("Search", string(a.cfg.Search.Backend)),
		section("Keys", keysHelp),
	}
	if !a.cfg.Provider.Ready() {
		entries = append([]chatEntry{section("Provider", "Press p in NORMAL mode to connect an OpenAI-compatible API")}, entries...)
	}
	return entries
}

// refreshIntro rebuilds the intro of a session nothing was sent to yet, so
// it reflects provider and search changes.
func (a *app) refreshIntro() {
	s := a.active
	if s.persisted || len(s.pending) > 0 {
		return
	}
	s.history, s.rows, s.scroll = nil, nil, 0
	for _, entry := range a.introEntries() {
		s.appendEntry(entry)
	}
}

func section(title, body string) chatEntry {
	return chatEntry{kind: core.UpdateInfo, tool: sectionEntry, text: title + "\n" + body}
}
