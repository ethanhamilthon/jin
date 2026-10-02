package ui

import (
	"slices"
	"strings"

	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/prompts"
	"jin/internal/sysprompt"
	"jin/internal/tools"
)

func (a *app) toolsLine() string {
	names := tools.Without(a.cfg.ToolsDisabled)
	if len(names) == 0 {
		return "no tools enabled"
	}
	return strings.Join(names, ", ")
}

func (a *app) introEntries() []chatEntry {
	entries := []chatEntry{logo(a.version), section("Tools", a.toolsLine()), section("Context", contextLines(core.ContextFiles(a.dir))), section("Hooks", hooksLine(hooks.Active(a.cfg.HooksDisabled))), promptsSection(a.promptsLine())}
	if sys := systemPromptLine(); sys != "" {
		entries = append(entries, section("System prompt", sys))
	}
	if !a.cfg.Provider.Ready() {
		entries = append(entries[:1:1], append([]chatEntry{section("Provider", "Type /provider to connect an OpenAI-compatible or Anthropic-compatible API")}, entries[1:]...)...)
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

// systemPromptLine says that the system prompt file is in use. Without the
// file the built-in prompts are used and the intro stays quiet.
func systemPromptLine() string {
	path, err := sysprompt.Path()
	if err != nil {
		return ""
	}
	if loaded, _ := sysprompt.Load(); !loaded.Custom {
		return ""
	}
	return "custom (" + shortPath(path) + ")"
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

func hooksLine(active []hooks.Hook) string {
	if len(active) == 0 {
		return "no hooks enabled"
	}
	names := make([]string, len(active))
	for i, hook := range active {
		names[i] = hook.Name
	}
	return strings.Join(names, ", ")
}

// promptsLine lists the enabled #prompts, system ones first.
func (a *app) promptsLine() string {
	infos, _ := prompts.ListInfo()
	var names []string
	for _, info := range infos {
		if !slices.Contains(a.cfg.PromptsDisabled, info.Name) {
			names = append(names, "#"+info.Name)
		}
	}
	if len(names) == 0 {
		return "no prompts enabled"
	}
	return strings.Join(names, ", ")
}
