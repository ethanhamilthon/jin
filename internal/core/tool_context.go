package core

import (
	"context"
	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

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
