package session

import (
	"strconv"

	"jin/internal/core"
	"jin/internal/diff"
)

// maxLines caps the output lines one tool result keeps.
const maxLines = 2000

func capLines(lines []diff.Line) []diff.Line {
	if len(lines) <= maxLines {
		return lines
	}
	hidden := len(lines) - maxLines
	note := diff.Line{Op: diff.Keep, Text: "… " + strconv.Itoa(hidden) + " more lines"}
	return append([]diff.Line{note}, lines[hidden:]...)
}

// applyUsage counts the tokens and cost of a request and notes the share of
// its input served from the provider's cache, when the provider says.
func (s *Session) applyUsage(u core.Update) {
	if !u.Usage.Known {
		return
	}
	if u.Usage.CacheKnown && u.Usage.Input > 0 {
		percent := min(100, u.Usage.CachedInput*100/u.Usage.Input)
		s.cache = &percent
	} else {
		s.cache = nil
	}
	s.usage.Add(u.Usage, u.Model, s.m.prices)
	s.persistUsage()
}
