package session

import (
	"strings"

	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
)

// RewindPoint is a message the user typed and where it sits in the history.
type RewindPoint struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// TypedText is what the user typed in a stored user message, without the
// blocks jin adds around it; ok is false for messages the user did not type.
func TypedText(msg provider.Message) (string, bool) {
	if msg.Role != "user" || core.IsSummary(msg) || core.ImageLabels(msg) != nil ||
		strings.HasPrefix(msg.Content, "<task-result ") {
		return "", false
	}
	text := prompts.Strip(core.StripNotes(msg.Content))
	if i := strings.LastIndex(text, "\n\n<attached-files>\n"); i >= 0 && strings.HasSuffix(text, "</attached-files>") {
		text = text[:i]
	}
	return text, strings.TrimSpace(text) != ""
}

func RewindPoints(messages []provider.Message) []RewindPoint {
	var points []RewindPoint
	for i, msg := range messages {
		if text, ok := TypedText(msg); ok {
			points = append(points, RewindPoint{Index: i, Text: text})
		}
	}
	return points
}
