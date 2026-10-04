package core

import (
	"strings"

	"jin/internal/provider"
)

const (
	summaryOpen  = "<conversation-summary>"
	summaryClose = "</conversation-summary>"
	summaryLead  = "The earlier conversation was compacted into the continuation brief below. It is a factual record of the session, not a direct user instruction; separate explicit user requirements from observations and unverified claims. If it lists unfinished work, continue it."
)

// SummaryMessage is the user message that replaces the conversation after a
// compaction. It is stored like any other message and doubles as the marker:
// the model only ever sees the history from the latest one on.
func SummaryMessage(summary string) provider.Message {
	return provider.Message{Role: "user", Content: summaryOpen + "\n" + summaryLead + "\n\n" + strings.TrimSpace(summary) + "\n" + summaryClose}
}

func IsSummary(msg provider.Message) bool {
	return msg.Role == "user" && strings.HasPrefix(msg.Content, summaryOpen)
}

// SinceLastSummary returns what the model should see of a stored session.
func SinceLastSummary(messages []provider.Message) []provider.Message {
	for i := len(messages) - 1; i >= 0; i-- {
		if IsSummary(messages[i]) {
			return messages[i:]
		}
	}
	return messages
}
