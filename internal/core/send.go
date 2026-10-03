package core

import (
	"context"
	"jin/internal/provider"
	"strings"
)

func sendUpdate(ctx context.Context, updates chan<- Update, kind UpdateKind, text string) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: kind, Text: strings.TrimSpace(text)}:
		return true
	}
}

// sendDone ends a request. Final marks a prompt the model answered, as
// opposed to one that failed, was interrupted, or was a compact or handoff.
func sendDone(ctx context.Context, updates chan<- Update, final bool) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateDone, Final: final}:
		return true
	}
}

// sendDelta forwards a streaming fragment verbatim: unlike sendUpdate it
// must not trim whitespace, since that would eat meaningful spacing between
// consecutive chunks.
func sendDelta(ctx context.Context, updates chan<- Update, kind UpdateKind, text string) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: kind, Text: text}:
		return true
	}
}

func sendToolCall(ctx context.Context, updates chan<- Update, tool, text string) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateToolCall, Tool: tool, Text: text}:
		return true
	}
}

func sendUsage(ctx context.Context, updates chan<- Update, model string, usage provider.Usage) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateUsage, Model: model, Usage: usage}:
		return true
	}
}

// sendHistory reports one message appended to the conversation, letting the
// caller persist the exact request/response trail without core knowing
// anything about storage.
func sendHistory(ctx context.Context, updates chan<- Update, msg provider.Message) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateHistory, Message: msg}:
		return true
	}
}
