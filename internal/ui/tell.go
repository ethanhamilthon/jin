package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

// tellRows draws a message the agent showed with tell_user: a marker in the
// accent color, then the text.
func tellRows(text string, width int) []chatRow {
	var rows []chatRow
	for i, line := range wrapChat(text, width-4) {
		marker := "  "
		if i == 0 {
			marker = "› "
		}
		rows = append(rows, chatRow{kind: core.UpdateTell, spans: []chatSpan{{text: marker, style: accent.Bold(true)}, {text: line, style: base}}})
	}
	return append(rows, chatRow{})
}

// suggestionKey handles Enter and → on an empty input that shows the agent's
// suggestion: Enter sends it, → puts it into the input to edit.
func (a *app) suggestionKey(ev *tcell.EventKey) bool {
	s := a.active
	if !s.showsSuggestion() || !ev.Pressed() || ev.Modifiers() != 0 {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEnter:
		text := s.suggestion
		s.suggestion = ""
		a.sendDraft(text)
		return true
	case tcell.KeyRight:
		insertClusters(&s.input, &s.cursor, s.suggestion)
		s.suggestion = ""
		return true
	}
	return false
}

// showsSuggestion reports whether the empty input offers the suggestion.
func (s *chatSession) showsSuggestion() bool {
	return s.suggestion != "" && len(s.input) == 0 && s.ready && !s.working && !s.bashInput()
}

const suggestionHint = "  · Enter send · → edit"

// drawSuggestion shows the suggestion greyed in the empty input, on one line.
func drawSuggestion(screen tcell.Screen, text string, top, width int) {
	room := width - 2 - displaywidth.String(suggestionHint)
	line := truncate(strings.Join(strings.Fields(text), " "), max(room, 1))
	put(screen, 2, top, line, muted)
	if room > 0 {
		put(screen, 2+displaywidth.String(line), top, suggestionHint, dim)
	}
}
