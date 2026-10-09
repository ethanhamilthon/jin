package core

import "strings"

// ExplainPrompt returns the labeled parts of a system prompt. A prompt made
// by BuildSystemPrompt in this process is explained as it was built; any
// other text is one part.
func ExplainPrompt(prompt string) []PromptPart {
	if parts, ok := built(prompt); ok {
		return parts
	}
	return []PromptPart{{"system prompt", strings.TrimSpace(prompt)}}
}
