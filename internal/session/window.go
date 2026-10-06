package session

import (
	"strconv"
	"strings"
)

// window is the model's context size in tokens: the catalogues first, then
// the setting models.window.<model id>, else 0 for unknown.
func (s *Session) window() int {
	if entry, _ := s.m.prices.Lookup(s.model); entry.MaxInputTokens > 0 {
		return entry.MaxInputTokens
	}
	value, _ := s.m.db.Setting("models.window." + s.model)
	tokens, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return max(0, tokens)
}

// noVision is true only when the catalogues say the model takes no images.
func (s *Session) noVision() bool {
	entry, ok := s.m.prices.Lookup(s.model)
	return ok && entry.VisionKnown && !entry.Vision
}
