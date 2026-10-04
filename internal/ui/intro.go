package ui

import (
	"strings"

	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/tools"
)

func (a *app) toolsLine() string {
	names := tools.Without(a.cfg.ToolsDisabled)
	if len(names) == 0 {
		return "no tools enabled"
	}
	return strings.Join(names, ", ")
}

func (a *app) introEntries() []chatEntry { return a.introEntriesAt(a.dir) }

func (a *app) introEntriesAt(dir string) []chatEntry {
	entries := []chatEntry{logo(a.version), section("Tools", a.toolsLine()), section("Context", contextLines(core.ContextFiles(dir))), section("Hooks", hooksLine(hooks.ActiveIn(dir, a.cfg.HooksDisabled, a.projectHooksTrustedAt(dir)))), promptsSection(a.promptsLine())}
	if sys := systemPromptLine(); sys != "" {
		entries = append(entries, section("System prompt", sys))
	}
	if a.latest != "" {
		entries = append(entries, updateSection(a.version, a.latest))
	}
	return entries
}

// refreshIntro rebuilds the intro of a session nothing was sent to yet, so
// it reflects provider changes.
func (a *app) refreshIntro() {
	s := a.active
	if s.persisted || len(s.pending) > 0 {
		return
	}
	s.history, s.rows, s.scroll = nil, nil, 0
	for _, entry := range a.introEntries() {
		s.appendEntry(entry)
	}
	s.syncLoading()
}

// promptsSection is the Prompts section. It has its own entry type so that
// the names of the prompts that still run commands can show a spinner.
func promptsSection(body string) chatEntry {
	return chatEntry{kind: core.UpdateInfo, tool: promptsEntry, text: "Prompts\n" + body}
}

// syncLoading copies the prompts that still load into the intro and redraws.
func (s *chatSession) syncLoading() {
	for i := range s.history {
		if s.history[i].tool == promptsEntry {
			s.history[i].pending = s.loadingPrompts()
		}
	}
	s.rebuildRows(s.width)
}

func section(title, body string) chatEntry {
	return chatEntry{kind: core.UpdateInfo, tool: sectionEntry, text: title + "\n" + body}
}
