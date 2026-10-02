package core

import (
	"context"

	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

const (
	imagesFromTools = "Attached image(s) from tool result:"
	noVisionNote    = "\n[The current model does not support images. The image is omitted from this request.]"
)

// runTools answers every call with a tool message. The pictures of the whole
// batch follow in one user message, but only once every call has an answer, so
// an interrupted batch never gets a user message between its tool messages.
func (a *Agent) runTools(work, ctx context.Context, request Request, calls []provider.ToolCall, history *[]provider.Message, updates chan<- Update) error {
	var pictures []provider.Image
	for _, call := range calls {
		if err := work.Err(); err != nil {
			return err
		}
		result, images := executeTool(a.toolContext(work, ctx, updates), ctx, call, a.registry, updates)
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

// executeTool runs one tool call. Pictures the tool returns come back
// separately, because a tool message can only carry text.
func executeTool(work, ctx context.Context, call provider.ToolCall, registry *tools.Registry, updates chan<- Update) (string, []provider.Image) {
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

func toMessageImages(images []tools.Image) []provider.Image {
	out := make([]provider.Image, len(images))
	for i, image := range images {
		out[i] = provider.Image{MimeType: image.MimeType, Data: image.Data}
	}
	return out
}

// ImageLabels returns the chat notes for a message made by imageMessage, and
// nothing for any other message.
func ImageLabels(msg provider.Message) []string {
	if msg.Role != "user" || msg.Content != imagesFromTools {
		return nil
	}
	labels := make([]string, len(msg.Images))
	for i, image := range msg.Images {
		labels[i] = "[" + image.Label() + "]"
	}
	return labels
}

// imageMessage carries the pictures of a batch of tool results to the model.
func imageMessage(images []provider.Image) provider.Message {
	return provider.Message{Role: "user", Content: imagesFromTools, Images: images}
}

// toolContext lets the ask_user and todo tools reach the UI through updates.
func (a *Agent) toolContext(work, ctx context.Context, updates chan<- Update) context.Context {
	work = tools.WithAsker(work, func(askCtx context.Context, questions []tools.Question) ([]string, error) {
		select {
		case <-a.answers:
		default:
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case updates <- Update{Kind: UpdateAsk, Questions: questions}:
		}
		select {
		case <-askCtx.Done():
			return nil, askCtx.Err()
		case answers := <-a.answers:
			return answers, nil
		}
	})
	return tools.WithTodoSink(work, func(items []todo.Item) {
		select {
		case <-ctx.Done():
		case updates <- Update{Kind: UpdateTodo, Todos: items}:
		}
	})
}
