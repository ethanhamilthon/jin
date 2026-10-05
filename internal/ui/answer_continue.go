package ui

import "jin/internal/core"

// continueLastAnswer shows the reply the agent gives after the background
// tasks note as the end of its final answer: the note stays hidden, so does
// the reasoning of the reply, and its text joins the answer above. It reports
// whether the delta is fully handled.
func (s *chatSession) continueLastAnswer(kind core.UpdateKind) bool {
	if kind == core.UpdateReasoning {
		return true
	}
	s.continueAnswer = false
	last := len(s.history) - 1
	if last < 0 || s.history[last].kind != core.UpdateAssistant {
		return false
	}
	s.history[last].text += "\n\n"
	s.openKind = core.UpdateAssistant
	s.rebuildRows(s.width)
	return false
}
