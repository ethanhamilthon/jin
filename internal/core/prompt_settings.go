package core

import (
	"strings"

	"jin/internal/sysprompt"
)

// SetSystemPrompt replaces the system prompt and notes its render time. Call it before Run.
func (a *Agent) SetSystemPrompt(text string) {
	a.mu.Lock()
	a.systemPrompt, a.renderedAt = text, a.now()
	a.mu.Unlock()
}

// SetSidePrompts sets the instructions for compaction and handoff. An empty
// text means the built-in default.
func (a *Agent) SetSidePrompts(compact, handoff string) {
	a.mu.Lock()
	a.compactText, a.handoffText = compact, handoff
	a.mu.Unlock()
}

func (a *Agent) compactPrompt() string {
	a.mu.Lock()
	text := a.compactText
	a.mu.Unlock()
	if strings.TrimSpace(text) != "" {
		return text
	}
	return sysprompt.Defaults().Compact
}

func (a *Agent) handoffPrompt() string {
	a.mu.Lock()
	text := a.handoffText
	a.mu.Unlock()
	if strings.TrimSpace(text) != "" {
		return text
	}
	return sysprompt.Defaults().Handoff
}
