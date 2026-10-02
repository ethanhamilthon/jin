package core

import (
	"context"
	"strings"

	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

type UpdateKind string

const (
	UpdateUser           UpdateKind = "user"
	UpdateAssistant      UpdateKind = "assistant"
	UpdateReasoning      UpdateKind = "reasoning"
	UpdateAssistantDelta UpdateKind = "assistant_delta"
	UpdateReasoningDelta UpdateKind = "reasoning_delta"
	UpdateUsage          UpdateKind = "usage"
	UpdateToolCall       UpdateKind = "tool_call"
	UpdateError          UpdateKind = "error"
	UpdateInfo           UpdateKind = "info"
	UpdateWorking        UpdateKind = "working"
	UpdateDone           UpdateKind = "done"
	UpdateHistory        UpdateKind = "history"
	UpdateCompacted      UpdateKind = "compacted"
	UpdateHandoff        UpdateKind = "handoff"
	UpdateAsk            UpdateKind = "ask"
	UpdateTodo           UpdateKind = "todo"
)

type Update struct {
	Kind    UpdateKind
	Text    string
	Tool    string
	Model   string
	Usage   provider.Usage
	Message provider.Message
	Final   bool
	// Questions is set on UpdateAsk, Todos on UpdateTodo.
	Questions []tools.Question
	Todos     []todo.Item
}

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
