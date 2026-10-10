package ui

import (
	"jin/internal/core"
	"jin/internal/prompts"
	"strings"
)

func (s *chatSession) send(text string) { s.sendFiles(text, text, "", nil) }

// sendFiles shows text in the chat and sends clean plus the attached-files
// block to the model, with the bodies of the named prompts.
func (s *chatSession) sendFiles(text, clean, block string, names []string) {
	s.closeOpenEntry()
	prompt := prompts.ExpandNames(clean, names, s.promptBodies)
	if block != "" {
		prompt += "\n\n" + block
	}
	request := core.Request{Prompt: prompt, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision()}
	if strings.TrimSpace(request.Prompt) != "" {
		// A command that runs now moves to the background, not in your way.
		request.Interactive = true
		s.agent.Expect()
	}
	s.pending = append(s.pending, request)
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: text})
	s.scroll = 0
	s.touch(text)
}
