package core

import (
	"context"
	_ "embed"
	"errors"
	"slices"
	"strings"

	"jin/internal/provider"
)

//go:embed handoff_prompt.md
var handoffPrompt string

// sideRequest asks the model something about the conversation without adding
// the question or the answer to it. The request is the whole history plus one
// extra user message, with the same model, effort and tools, so a provider
// with prompt caching bills only that last message at the full rate.
func (a *Agent) sideRequest(work context.Context, request Request, history []provider.Message, instruction string) (string, provider.Usage, error) {
	if request.Model == "" {
		return "", provider.Usage{}, errNoModel
	}
	messages := append(slices.Clone(history), provider.Message{Role: "user", Content: strings.TrimSpace(instruction)})
	response, err := a.client.Stream(work, request.Model, request.Effort, messages, a.registry.SchemaJSON(), func(provider.StreamEvent) {})
	if err != nil {
		return "", provider.Usage{}, err
	}
	text := strings.TrimSpace(response.Message.Content)
	if text == "" {
		return "", response.Usage, errors.New("the model returned no text")
	}
	return text, response.Usage, nil
}

// handoff writes a brief for a new session and leaves this one untouched.
func (a *Agent) handoff(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) error {
	if len(*history) < 2 {
		return errors.New("nothing to hand off")
	}
	brief, usage, err := a.sideRequest(work, request, *history, handoffPrompt)
	if usage.Known {
		sendUsage(ctx, updates, request.Model, usage)
	}
	if err != nil {
		return err
	}
	sendUpdate(ctx, updates, UpdateHandoff, brief)
	return nil
}
