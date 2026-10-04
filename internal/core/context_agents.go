package core

const agentsHeading = "AGENTS.md:\n"

// agentsParts is the AGENTS.md section as one part per file, so that
// joining them with a blank line gives agentsHeading + renderContext(files).
func agentsParts(files []ContextFile) []PromptPart {
	if len(files) == 0 {
		return []PromptPart{{"AGENTS.md", agentsHeading + renderContext(nil)}}
	}
	parts := make([]PromptPart, len(files))
	for i, f := range files {
		parts[i] = PromptPart{"AGENTS.md " + f.Path, renderContext([]ContextFile{f})}
	}
	parts[0].Text = agentsHeading + parts[0].Text
	return parts
}
