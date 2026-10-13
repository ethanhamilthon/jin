package ui

import (
	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/prompts"
	"slices"
	"strings"
)

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
		if hook.Project {
			names[i] += " (project)"
		}
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
