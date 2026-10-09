package core

import (
	"jin/internal/wire"
	"strings"

	"jin/internal/provider"
)

const (
	summaryOpen  = wire.SummaryOpen
	summaryClose = wire.SummaryClose
	summaryLead  = wire.SummaryLead
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
