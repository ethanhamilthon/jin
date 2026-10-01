package core

import (
	"context"

	"jin/internal/provider"
	"jin/internal/tools"
)

func executeTool(work, ctx context.Context, call provider.ToolCall, registry *tools.Registry, updates chan<- Update) string {
	tool, ok := registry.Get(call.Function.Name)
	if !ok || call.ID == "" {
		return "Unsupported tool call"
	}
	summary, ok := tool.Summary(call.Function.Arguments)
	if !ok {
		return "Invalid " + call.Function.Name + " tool arguments"
	}
	sendToolCall(ctx, updates, call.Function.Name, summary)
	result, err := tool.Run(work, call.Function.Arguments)
	if err != nil {
		return "Error: " + err.Error()
	}
	return result
}
