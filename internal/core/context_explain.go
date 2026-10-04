package core

import (
	"strings"

	"jin/internal/provider"
)

// ExplainPrompt cuts a rendered system prompt back into labeled parts. The
// fixed texts and the AGENTS.md files of dir are found by their text; hooks
// are matched in order by their body. What cannot be matched (a hook with
// commands in it, an AGENTS.md changed since the start) stays in one part.
func ExplainPrompt(prompt, dir string, hooks []PromptPart) []PromptPart {
	head, tail, found := strings.Cut(prompt, provider.CacheBreak)
	if !found {
		return []PromptPart{{"system prompt", prompt}}
	}
	head = strings.TrimSuffix(head, "\n\n")
	files := agentsParts(ContextFiles(dir))
	section := joinParts(files)
	var agents []PromptPart
	if rest, ok := strings.CutSuffix(head, section); ok {
		head, agents = strings.TrimSuffix(rest, "\n\n"), files
	} else if i := strings.LastIndex(head, "\n\n"+agentsHeading); i >= 0 {
		head, agents = head[:i], []PromptPart{{"AGENTS.md (changed since start)", head[i+2:]}}
	}
	parts := explainFront(head, hooks)
	parts = append(parts, agents...)
	parts = append(parts, PromptPart{"cache break", provider.CacheBreak})
	if tail = strings.TrimSpace(tail); tail != "" {
		parts = append(parts, PromptPart{"environment", tail})
	}
	return parts
}

func joinParts(parts []PromptPart) string {
	texts := make([]string, len(parts))
	for i, part := range parts {
		texts[i] = part.Text
	}
	return strings.Join(texts, "\n\n")
}

// explainFront splits what comes before the AGENTS.md files: the system
// text, the docs pointer, the async instructions and the hooks.
func explainFront(front string, hooks []PromptPart) []PromptPart {
	docs := strings.TrimSpace(docsPrompt)
	before, after, found := strings.Cut(front, docs)
	if !found {
		return []PromptPart{{"system text", front}}
	}
	var parts []PromptPart
	if before = strings.TrimSpace(before); before != "" {
		parts = append(parts, PromptPart{"system text", before})
	}
	parts = append(parts, PromptPart{"jin docs", docs})
	after = strings.TrimPrefix(after, "\n\n")
	if rest, ok := strings.CutPrefix(after, strings.TrimSpace(asyncPrompt)); ok {
		parts = append(parts, PromptPart{"async instructions", strings.TrimSpace(asyncPrompt)})
		after = strings.TrimPrefix(rest, "\n\n")
	}
	for _, hook := range hooks {
		rest, ok := strings.CutPrefix(after, hook.Text)
		if !ok {
			break
		}
		parts = append(parts, hook)
		after = strings.TrimPrefix(rest, "\n\n")
	}
	if after != "" {
		parts = append(parts, PromptPart{"hooks (with command output)", after})
	}
	return parts
}
