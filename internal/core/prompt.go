package core

import (
	_ "embed"
	"strings"

	"jin/internal/tools"
)

//go:embed docs_prompt.md
var docsPrompt string

//go:embed async_prompt.md
var asyncPrompt string

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

// BuildSystemPrompt joins the parts, in this order: the system text, the
// tool list, the hooks, the jin docs pointer, the async instructions and the
// AGENTS.md files. The docs pointer is always there. The async instructions
// are there only when the agent can use them: it needs the bash tool to run
// `jin async` and a session id to give to it. No command is run and no
// placeholder is replaced here, so text from an AGENTS.md can never act.
func BuildSystemPrompt(in PromptInput) string {
	var parts []string
	if system := strings.TrimSpace(in.System); system != "" {
		parts = append(parts, system)
	}
	parts = append(parts, strings.TrimRight(renderTools(in.ToolNames), "\n"))
	for _, hook := range in.Hooks {
		if hook = strings.TrimSpace(hook); hook != "" {
			parts = append(parts, hook)
		}
	}
	parts = append(parts, strings.TrimSpace(docsPrompt))
	if in.SessionID != "" && hasTool(in.ToolNames, "bash") {
		parts = append(parts, strings.TrimSpace(asyncPrompt)+"\n- Your session id: "+in.SessionID)
	}
	parts = append(parts, "AGENTS.md:\n"+renderContext(ContextFiles(in.Dir)))
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

func renderTools(names []string) string {
	if len(names) == 0 {
		return "Tools: none. You cannot read files or run commands; answer from what the user tells you.\n"
	}
	var b strings.Builder
	b.WriteString("Tools:\n")
	for _, name := range names {
		b.WriteString("- " + name + ": " + tools.Describe(name) + "\n")
	}
	return b.String()
}
