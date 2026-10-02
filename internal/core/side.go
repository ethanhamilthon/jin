package core

import (
	"context"
	"errors"
	"slices"
	"strings"

	"jin/internal/provider"
)

const sideRetryReminder = "\n\nImportant: Reply with text only. Do not invoke any tools or function calls."

// sideRequest asks the model something about the conversation without adding
// the question or the answer to it. The request is the whole history plus one
// extra user message, with the same model, effort and tools, so a provider
// with prompt caching bills only that last message at the full rate. If the
// model returns a tool call instead of text, it retries once with a reminder.
func (a *Agent) sideRequest(work, ctx context.Context, request Request, history []provider.Message, instruction string, updates chan<- Update) (string, provider.Usage, error) {
	if request.Model == "" {
		return "", provider.Usage{}, errNoModel
	}
	for attempt := 0; attempt < 2; attempt++ {
		prompt := strings.TrimSpace(instruction)
		if attempt > 0 {
			prompt += sideRetryReminder
		}
		messages := append(slices.Clone(history), provider.Message{Role: "user", Content: prompt})
		response, err := a.client.Stream(work, request.Model, request.Effort, messages, a.registry.SchemaJSON(), func(provider.StreamEvent) {})
		if err != nil {
			return "", provider.Usage{}, err
		}
		text := strings.TrimSpace(response.Message.Content)
		if len(response.Message.ToolCalls) == 0 && text != "" {
			return text, response.Usage, nil
		}
		if attempt == 0 && work.Err() == nil {
			if response.Usage.Known {
				sendUsage(ctx, updates, request.Model, response.Usage)
			}
			continue
		}
		return "", response.Usage, errors.New("the model returned no text")
	}
	return "", provider.Usage{}, errors.New("the model returned no text")
}

// handoff writes a brief for a new session and leaves this one untouched.
func (a *Agent) handoff(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) error {
	if len(*history) < 2 {
		return errors.New("nothing to hand off")
	}
	brief, usage, err := a.sideRequest(work, ctx, request, *history, a.handoffPrompt(), updates)
	if usage.Known {
		sendUsage(ctx, updates, request.Model, usage)
	}
	if err != nil {
		return err
	}
	sendUpdate(ctx, updates, UpdateHandoff, brief)
	return nil
}
