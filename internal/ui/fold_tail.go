package ui

import (
	"github.com/clipperhouse/displaywidth"

	"jin/internal/core"
)

// tailRows is the line closing the chat: the Ctrl+O hint, or while the agent
// works behind a fold, a spinner with the last thing it did.
func (a *app) tailRows(s *chatSession) []chatRow {
	if !s.hasFoldable() {
		return nil
	}
	spans := []chatSpan{{text: s.fold.hint(), style: dim}}
	if s.working && s.fold != foldAll {
		spans = a.workingSpans(s)
	}
	var rows []chatRow
	if len(s.rows) > 0 && !s.rows[len(s.rows)-1].blank() {
		rows = append(rows, chatRow{})
	}
	return append(rows, chatRow{kind: core.UpdateInfo, spans: spans})
}

func (a *app) workingSpans(s *chatSession) []chatSpan {
	frame := spinnerFrames[a.frame%len(spinnerFrames)]
	spans := []chatSpan{{text: frame, style: base.Foreground(colorAmber)}, {text: " working...", style: muted}}
	hint := " · " + s.fold.hint()
	used := displaywidth.String(frame) + len(" working...") + len(hint) + 2
	if tool, text, ok := s.lastAction(); ok {
		room := max(8, s.width-used-displaywidth.String(tool)-4)
		spans = append(spans,
			chatSpan{text: "  " + tool, style: base.Foreground(toolAccent(tool)).Bold(true)},
			chatSpan{text: " " + truncate(firstLine(text), room), style: base.Foreground(colorArgument)})
	}
	return append(spans, chatSpan{text: hint, style: dim})
}

func (s *chatSession) hasFoldable() bool {
	for _, entry := range s.history {
		if entry.kind == core.UpdateToolCall || entry.kind == core.UpdateReasoning {
			return true
		}
	}
	return false
}

// lastAction is the newest hidden thing the agent did since the user's last
// message: a tool call, or in the strictest fold also its reasoning.
func (s *chatSession) lastAction() (name, text string, ok bool) {
	for i := len(s.history) - 1; i >= 0; i-- {
		entry := s.history[i]
		switch {
		case entry.kind == core.UpdateUser:
			return "", "", false
		case entry.kind == core.UpdateToolCall:
			return entry.tool, entry.text, true
		case entry.kind == core.UpdateReasoning && s.fold == foldMessages && entry.text != "":
			return "thinking", entry.text, true
		}
	}
	return "", "", false
}
