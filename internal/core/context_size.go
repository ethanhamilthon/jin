package core

import "jin/internal/provider"

// SetContextSize restores the last reported context size before Run starts.
func (a *Agent) SetContextSize(tokens int) { a.size = max(0, tokens) }

// startSize sets up the estimate when Run starts. The restored size may
// come from a pruned history while the stored one is complete, so the
// larger of it and an estimate of the whole history wins.
func (a *Agent) startSize(history []provider.Message) {
	restored := a.size
	a.reseed(history)
	a.size = max(a.size, restored)
}

// reported records the context size a provider response reported; it covers
// the history as it is once the response is appended.
func (a *Agent) reported(usage provider.Usage, history []provider.Message) {
	if usage.Known {
		a.size, a.mark = usage.Input+usage.Output, len(history)
	}
}

// reseed estimates the whole request from scratch: the history plus the
// tool schemas.
func (a *Agent) reseed(history []provider.Message) {
	a.size, a.mark = len(a.registry.SchemaJSON())/4, 0
	a.size, a.mark = a.contextSize(history), len(history)
}

// contextSize is the last reported size plus an estimate of every message
// appended since that report.
func (a *Agent) contextSize(history []provider.Message) int {
	n := a.size
	for _, msg := range history[min(a.mark, len(history)):] {
		n += messageTokens(msg)
	}
	return n
}

// messageTokens is a cheap estimate: four bytes per token.
func messageTokens(msg provider.Message) int {
	n := len(msg.Content) + len(msg.ReasoningContent)
	for _, call := range msg.ToolCalls {
		n += len(call.Function.Name) + len(call.Function.Arguments)
	}
	return n / 4
}
