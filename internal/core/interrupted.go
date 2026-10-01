package core

import "jin/internal/provider"

const interruptedToolResult = "Interrupted; tool result unknown. The tool may have already run. Call it again only if needed."

// InterruptedToolMessages answers every tool call that never got a result,
// so the history stays valid for the next request.
func InterruptedToolMessages(messages []provider.Message) []provider.Message {
	pending := make(map[string]bool)
	var order []string
	for _, msg := range messages {
		for _, call := range msg.ToolCalls {
			if call.ID != "" && !pending[call.ID] {
				order = append(order, call.ID)
				pending[call.ID] = true
			}
		}
		if msg.Role == "tool" {
			delete(pending, msg.ToolCallID)
		}
	}
	var results []provider.Message
	for _, id := range order {
		if pending[id] {
			results = append(results, provider.Message{Role: "tool", ToolCallID: id, Content: interruptedToolResult})
		}
	}
	return results
}
