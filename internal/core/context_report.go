package core

import (
	"sort"

	"jin/internal/provider"
)

// EstimateTokens is the cheap estimate used everywhere: four bytes per token.
func EstimateTokens(bytes int) int { return bytes / 4 }

// ConversationBytes is the size of the text of the messages.
func ConversationBytes(messages []provider.Message) int {
	n := 0
	for _, msg := range messages {
		n += len(msg.Content) + len(msg.ReasoningContent)
		for _, call := range msg.ToolCalls {
			n += len(call.Function.Name) + len(call.Function.Arguments)
		}
	}
	return n
}

// EffectiveMessages is the conversation as the agent holds it. Old tool
// results are shortened once the request reaches the prune threshold, and
// the stored messages only grow, so a stored size past the threshold means
// the live history is pruned. baseBytes is the system prompt plus the tool
// schemas.
func EffectiveMessages(messages []provider.Message, baseBytes, window int) []provider.Message {
	if !needsPrune(EstimateTokens(baseBytes+ConversationBytes(messages)), window) {
		return messages
	}
	pruned, _ := pruneToolResults(messages, pruneKeepTurns)
	return pruned
}

// ToolResult is one tool result with the call that produced it.
type ToolResult struct {
	Call  provider.ToolCall
	Bytes int
}

// LargestToolResults returns the n biggest tool results, biggest first.
func LargestToolResults(messages []provider.Message, n int) []ToolResult {
	calls := map[string]provider.ToolCall{}
	var results []ToolResult
	for _, msg := range messages {
		for _, call := range msg.ToolCalls {
			calls[call.ID] = call
		}
		if msg.Role == "tool" {
			results = append(results, ToolResult{Call: calls[msg.ToolCallID], Bytes: len(msg.Content)})
		}
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Bytes > results[j].Bytes })
	return results[:min(n, len(results))]
}

// SystemPrompt is the system prompt the agent was given.
func (a *Agent) SystemPrompt() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.systemPrompt
}

// ToolSchemaBytes is the size of the tool schemas sent with every request.
func (a *Agent) ToolSchemaBytes() int { return len(a.registry.SchemaJSON()) }
