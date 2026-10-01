package ui

import (
	"strings"

	"jin/internal/core"
)

func (a *app) toolsLine() string {
	names := a.registry.Names()
	for i, name := range names {
		if name == "websearch" {
			names[i] += " (" + string(a.cfg.Search.Backend) + ")"
		}
	}
	return strings.Join(names, ", ")
}

func (a *app) introEntries() []chatEntry {
	entries := []chatEntry{section("Tools", a.toolsLine()), section("Context", contextLines(core.ContextFiles(a.dir)))}
	if !a.cfg.Provider.Ready() {
		entries = append([]chatEntry{section("Provider", "Press Space, open Model & Provider, then Provider, to connect an OpenAI-compatible API")}, entries...)
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

func contextLines(files []core.ContextFile) string {
	if len(files) == 0 {
		return "no AGENTS.md found"
	}
	lines := make([]string, len(files))
	for i, f := range files {
		lines[i] = shortPath(f.Path) + " (" + string(f.Kind) + ")"
	}
	return strings.Join(lines, "\n")
}
