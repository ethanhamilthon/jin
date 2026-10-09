package core

import (
	"jin/internal/provider"
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
	// UpdateReset voids the deltas of a failed attempt that is retried; the
	// usage that attempt reported follows as an UpdateUsage.
	UpdateReset UpdateKind = "reset"
	// UpdateToolResult carries what a bash, edit or write call produced.
	UpdateToolResult UpdateKind = "tool_result"
	// UpdateTaken says the agent took a request that gets no turn and no
	// UpdateDone of its own: a prompt folded into the running turn, or one it
	// dropped. A front end counting requests in flight counts it as done.
	UpdateTaken UpdateKind = "taken"
)

type Update struct {
	Kind    UpdateKind
	Text    string
	Tool    string
	Model   string
	Usage   provider.Usage
	Message provider.Message
	Final   bool
	// Questions is set on UpdateAsk.
	Questions []tools.Question
	// CallID and Changes are set on UpdateToolResult; Changes lists the
	// files an edit or write call changed.
	CallID  string
	Changes []tools.Change
	// Image is the picture an UpdateInfo shows; Text is its label.
	Image *provider.Image
}
