package core

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"jin/internal/provider"
)

// Title asks a model to name the conversation. Like a side request it sends
// the history plus the prompt as one extra user message, with the same tool
// schema, and returns the text of the answer. It makes one attempt.
func Title(ctx context.Context, client *provider.Client, model, effort string, history []provider.Message, tools json.RawMessage, prompt string) (string, error) {
	if model == "" {
		return "", errNoModel
	}
	messages := append(slices.Clone(history), provider.Message{Role: "user", Content: strings.TrimSpace(prompt)})
	response, err := client.Stream(ctx, model, effort, messages, tools, func(provider.StreamEvent) {})
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(response.Message.Content)
	if text == "" {
		return "", errors.New("the model returned no title")
	}
	return text, nil
}
