package core

import (
	_ "embed"
	"strings"

	"jin/internal/provider"
)

//go:embed docs_prompt.md
var docsPrompt string

//go:embed async_prompt.md
var asyncPrompt string

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
}

// BuildSystemPrompt joins the parts, most stable first: the system text, a
// line when there are no tools (the tool schemas describe the others), the
// jin docs pointer and the async instructions; then the hooks and the
// AGENTS.md files; then provider.CacheBreak; then the environment and
// the session id. The docs pointer is always there. The async instructions
// are there only when the agent can use them: it needs the bash tool to run
// `jin async` and a session id to give to it. No command is run and no
// placeholder is replaced here, so text from an AGENTS.md can never act.
func BuildSystemPrompt(in PromptInput) string {
	var parts []string
	system := strings.TrimSpace(in.System)
	if system != "" {
		parts = append(parts, system)
	}
	if len(in.ToolNames) == 0 {
		parts = append(parts, noTools)
	}
	parts = append(parts, strings.TrimSpace(docsPrompt))
	if in.SessionID != "" && hasTool(in.ToolNames, "bash") {
		parts = append(parts, strings.TrimSpace(asyncPrompt))
	}
	for _, hook := range in.Hooks {
		if hook = strings.TrimSpace(hook); hook != "" {
			parts = append(parts, hook)
		}
	}
	parts = append(parts, "AGENTS.md:\n"+renderContext(ContextFiles(in.Dir)), provider.CacheBreak)
	if tail := sessionTail(in, system); tail != "" {
		parts = append(parts, tail)
	}
	return strings.Join(parts, "\n\n")
}

func hasTool(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
