package core

import (
	"context"
	"encoding/json"
	"sync"

	"jin/internal/provider"
)

type callOutcome struct {
	result string
	images []provider.Image
}

// readOnly reports whether the call only reads, so it may run next to others.
func readOnly(call provider.ToolCall) bool {
	switch call.Function.Name {
	case "read":
		return true
	case "todo":
		var args struct {
			Items json.RawMessage `json:"items"`
		}
		return json.Unmarshal([]byte(call.Function.Arguments), &args) == nil && (len(args.Items) == 0 || string(args.Items) == "null")
	}
	return false
}

// nextGroup is the longest run of read-only calls at the front, or the first
// call alone when it is not read-only.
func nextGroup(calls []provider.ToolCall) []provider.ToolCall {
	end := 1
	if readOnly(calls[0]) {
		for end < len(calls) && readOnly(calls[end]) {
			end++
		}
	}
	return calls[:end]
}

// runGroup executes the calls at the same time and returns the outcomes in
// call order.
func (a *Agent) runGroup(work, ctx context.Context, group []provider.ToolCall, updates chan<- Update) []callOutcome {
	outcomes := make([]callOutcome, len(group))
	var wait sync.WaitGroup
	for i, call := range group {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, images := executeTool(a.backgroundContext(a.toolContext(work, ctx, updates)), ctx, call, a.registry, updates)
			outcomes[i] = callOutcome{result, images}
		}()
	}
	wait.Wait()
	return outcomes
}
