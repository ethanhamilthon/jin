package headless

import (
	"strings"

	"jin/internal/core"
	"jin/internal/session"
)

// ourTurnStart is the first entry after our own message, or len(entries) when
// the session does not hold that message.
func ourTurnStart(entries []session.Entry, before int, prompt string) int {
	if before > len(entries) {
		before = len(entries)
	}
	for i := len(entries) - 1; i >= before; i-- {
		if entries[i].Kind == core.UpdateUser && strings.TrimSpace(entries[i].Text) == strings.TrimSpace(prompt) {
			return i + 1
		}
	}
	return len(entries)
}

// answerEntries picks the assistant entries that answer our own message: the
// entries after it, so work another client started in between stays out. An
// empty result means the session added no answer of ours.
func answerEntries(entries []session.Entry, before int, prompt string) []session.Entry {
	start := ourTurnStart(entries, before, prompt)
	if start >= len(entries) || start < before {
		return nil
	}
	var out []session.Entry
	for _, entry := range entries[start:] {
		if entry.Kind == core.UpdateAssistant && strings.TrimSpace(entry.Text) != "" {
			out = append(out, entry)
		}
	}
	return out
}
