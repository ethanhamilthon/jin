package core

import (
	"context"

	"jin/internal/provider"
	"jin/internal/tools"
)

const (
	imagesFromTools = "Attached image(s) from tool result:"
	noVisionNote    = "\n[The current model does not support images. The image is omitted from this request.]"
)

// runTools answers every call with a tool message, in call order. A run of
// read-only calls executes concurrently; every other call runs alone. The
// pictures of the whole batch follow in one user message, but only once every
// call has an answer, so an interrupted batch never gets a user message
// between its tool messages.
func (a *Agent) runTools(work, ctx context.Context, request Request, calls []provider.ToolCall, history *[]provider.Message, updates chan<- Update) error {
	var pictures []provider.Image
	for start := 0; start < len(calls); {
		if err := work.Err(); err != nil {
			return err
		}
		group := nextGroup(calls[start:])
		for i, done := range a.runGroup(work, ctx, group, updates) {
			call := group[i]
			result, images := done.result, done.images
			if request.NoVision && len(images) > 0 {
				result, images = result+noVisionNote, nil
			}
			pictures = append(pictures, images...)
			toolMessage := provider.Message{Role: "tool", ToolCallID: call.ID, Content: result}
			*history = append(*history, toolMessage)
			if !sendHistory(ctx, updates, toolMessage) {
				return ctx.Err()
			}
		}
		start += len(group)
	}
	if len(pictures) == 0 {
		return nil
	}
	attached := imageMessage(pictures)
	*history = append(*history, attached)
	if !sendHistory(ctx, updates, attached) {
		return ctx.Err()
	}
	for _, text := range ImageLabels(attached) {
		sendUpdate(ctx, updates, UpdateInfo, text)
	}
	return nil
}

// executeTool runs one tool call and reports its result to the UI. Pictures
// the tool returns come back separately, because a tool message can only
// carry text.
func executeTool(work, ctx context.Context, call provider.ToolCall, registry *tools.Registry, updates chan<- Update) (string, []provider.Image) {
	var changes []tools.Change
	work = tools.WithChangeSink(work, func(change tools.Change) { changes = append(changes, change) })
	result, images := runTool(work, ctx, call, registry, updates)
	if showsResult(call.Function.Name) {
		select {
		case <-ctx.Done():
		case updates <- Update{Kind: UpdateToolResult, Tool: call.Function.Name, CallID: call.ID, Text: result, Changes: changes}:
		}
	}
	return result, images
}

func showsResult(tool string) bool { return tool == "bash" || tool == "edit" || tool == "write" }

func runTool(work, ctx context.Context, call provider.ToolCall, registry *tools.Registry, updates chan<- Update) (string, []provider.Image) {
	tool, ok := registry.Get(call.Function.Name)
	if !ok || call.ID == "" {
		return "Unsupported tool call", nil
	}
	summary, ok := tool.Summary(call.Function.Arguments)
	if !ok {
		return "Invalid " + call.Function.Name + " tool arguments", nil
	}
	sendToolCall(ctx, updates, call.Function.Name, summary)
	if withImages, ok := tool.(tools.ImageTool); ok {
		text, images, err := withImages.RunImages(work, call.Function.Arguments)
		if err != nil {
			return "Error: " + err.Error(), nil
		}
		return text, toMessageImages(images)
	}
	result, err := tool.Run(work, call.Function.Arguments)
	if err != nil {
		return "Error: " + err.Error(), nil
	}
	return result, nil
}
