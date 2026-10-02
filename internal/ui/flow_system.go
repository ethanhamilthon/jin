package ui

import (
	"jin/internal/sysprompt"
)

// openSystemPromptFlow is /system-prompt: it opens ~/.jin/system-prompt.md in
// the editor, creating it from the built-in prompts when it is not there yet.
// The file holds the system prompt, the compaction prompt and the handoff
// prompt. Changes reach new sessions.
func (a *app) openSystemPromptFlow() {
	path, err := sysprompt.Create()
	if err != nil {
		a.report(err)
		return
	}
	a.report(a.editFile(path, func() error { return nil }))
}
