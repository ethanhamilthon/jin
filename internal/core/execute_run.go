package core

import (
	"context"

	"jin/internal/provider"
)

// runGroup executes the calls at the same time and returns the outcomes in
// call order. Results reach the UI in call order too, each as soon as it and
// every call before it have finished.
func (a *Agent) runGroup(work, ctx context.Context, group []provider.ToolCall, updates chan<- Update) []callOutcome {
	outcomes := make([]callOutcome, len(group))
	done := make([]chan struct{}, len(group))
	for i, call := range group {
		done[i] = make(chan struct{})
		go func() {
			defer close(done[i])
			outcomes[i] = executeTool(a.backgroundContext(a.toolContext(work, ctx, updates)), ctx, call, a.registry, updates)
		}()
	}
	for i, call := range group {
		<-done[i]
		if showsResult(call.Function.Name) {
			select {
			case <-ctx.Done():
			case updates <- Update{Kind: UpdateToolResult, Tool: call.Function.Name, CallID: call.ID, Text: outcomes[i].result, Changes: outcomes[i].changes}:
			}
		}
	}
	return outcomes
}
