package startup

import (
	"slices"
	"strings"

	"jin/internal/core"
	"jin/internal/dyn"
	"jin/internal/provider"
)

// systemParts labels the pieces of the filled system prompt file: its own
// text, the output of each command, and the cache break line.
func systemParts(segments []dyn.Segment) []core.PromptPart {
	var parts []core.PromptPart
	for _, segment := range segments {
		if segment.Command != "" {
			parts = append(parts, core.PromptPart{Name: "command: " + oneLine(segment.Command), Text: segment.Text})
			continue
		}
		before, after, found := strings.Cut(segment.Text, provider.CacheBreak)
		for found {
			parts = append(parts, textPart(before), core.PromptPart{Name: "cache break", Text: provider.CacheBreak})
			before, after, found = strings.Cut(after, provider.CacheBreak)
		}
		parts = append(parts, textPart(before))
	}
	return keep(parts)
}

func textPart(text string) core.PromptPart { return core.PromptPart{Name: "system text", Text: text} }

func keep(parts []core.PromptPart) []core.PromptPart {
	return slices.DeleteFunc(parts, func(p core.PromptPart) bool { return p.Text == "" })
}

func oneLine(command string) string {
	command = strings.Join(strings.Fields(command), " ")
	if len(command) > 48 {
		command = command[:47] + "…"
	}
	return command
}
