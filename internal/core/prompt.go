package core

import "strings"

// PromptPart is one labeled piece of the system prompt.
type PromptPart struct {
	Name string
	Text string
}

// Bytes is the size of the part.
func (p PromptPart) Bytes() int { return len(p.Text) }

// BuildSystemPrompt makes the system prompt from the pieces of the filled
// system prompt file. Jin adds nothing of its own: the prompt is that file
// with its commands run, and the pieces say which text each command made.
func BuildSystemPrompt(parts []PromptPart) string {
	prompt := strings.TrimSpace(joinParts(parts))
	remember(prompt, parts)
	return prompt
}

func joinParts(parts []PromptPart) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Text)
	}
	return b.String()
}
