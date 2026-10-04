package core

import "sync"

// maxBuilt bounds the memory of built prompts; one entry is a few KB.
const maxBuilt = 256

var builtPrompts = struct {
	sync.Mutex
	parts map[string][]PromptPart
}{parts: map[string][]PromptPart{}}

// remember keeps the parts of a built prompt, so that ExplainPrompt can name
// them later even when a hook's commands or an AGENTS.md changed since.
func remember(prompt string, parts []PromptPart) {
	builtPrompts.Lock()
	defer builtPrompts.Unlock()
	if len(builtPrompts.parts) >= maxBuilt {
		clear(builtPrompts.parts)
	}
	builtPrompts.parts[prompt] = parts
}

func built(prompt string) ([]PromptPart, bool) {
	builtPrompts.Lock()
	defer builtPrompts.Unlock()
	parts, ok := builtPrompts.parts[prompt]
	return parts, ok
}
