package core

import (
	_ "embed"
	"strconv"
	"strings"

	"jin/internal/provider"
)

//go:embed docs_prompt.md
var docsPrompt string

const noTools = "Tools: none. You cannot read files or run commands; answer from what the user tells you."

// PromptInput is everything the final system prompt is made of. The texts
// are already filled in (commands run); building the prompt only joins them.
type PromptInput struct {
	// System is the system prompt text, from ~/.jin/system-prompt.md or the default.
	System string
	// Dir is the working directory, used to find the AGENTS.md files.
	Dir       string
	SessionID string
	ToolNames []string
	// Hooks are the bodies of the enabled hooks, in the order they go in.
	Hooks []string
	// OS is the operating system of the user; empty means this machine's.
	OS string
}

// PromptPart is one labeled piece of the system prompt.
type PromptPart struct {
	Name string
	Text string
}

// Bytes is the size of the part.
func (p PromptPart) Bytes() int { return len(p.Text) }

// BuildSystemPrompt joins the parts of SystemPromptParts with a blank line.
func BuildSystemPrompt(in PromptInput) string {
	parts := SystemPromptParts(in)
	texts := make([]string, len(parts))
	for i, part := range parts {
		texts[i] = part.Text
	}
	prompt := strings.Join(texts, "\n\n")
	remember(prompt, parts)
	return prompt
}

// SystemPromptParts returns the prompt as labeled parts, most stable first:
// the system text, a line when there are no tools (the tool schemas
// describe the others) and the jin docs pointer; then the hooks and the
// AGENTS.md files; then provider.CacheBreak; then the environment and the
// session id. The docs pointer is always there. No command is run and no placeholder is replaced here, so text from an
// AGENTS.md can never act.
func SystemPromptParts(in PromptInput) []PromptPart {
	var parts []PromptPart
	system := strings.TrimSpace(in.System)
	if system != "" {
		parts = append(parts, PromptPart{"system text", system})
	}
	if len(in.ToolNames) == 0 {
		parts = append(parts, PromptPart{"no tools notice", noTools})
	}
	parts = append(parts, PromptPart{"jin docs", strings.TrimSpace(docsPrompt)})
	for i, hook := range in.Hooks {
		if hook = strings.TrimSpace(hook); hook != "" {
			parts = append(parts, PromptPart{"hook " + strconv.Itoa(i+1), hook})
		}
	}
	parts = append(parts, agentsParts(ContextFiles(in.Dir))...)
	parts = append(parts, PromptPart{"cache break", provider.CacheBreak})
	if tail := sessionTail(in, system); tail != "" {
		parts = append(parts, PromptPart{"environment", tail})
	}
	return parts
}
