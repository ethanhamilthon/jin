package core

import (
	"context"
	"errors"

	"jin/internal/provider"
)

func (a *Agent) perform(work, ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) error {
	switch request.Kind {
	case RequestCompact:
		return a.compact(work, ctx, request, history, updates)
	case RequestHandoff:
		return a.handoff(work, ctx, request, history, updates)
	}
	if err := a.compactIfNeeded(work, ctx, request, history, updates); err != nil {
		return err
	}
	a.refreshSystem(work, *history)
	userMessage := provider.Message{Role: "user", Content: a.takeRefreshNote() + request.Prompt, Images: request.Images}
	*history = append(*history, userMessage)
	if !sendHistory(ctx, updates, userMessage) {
		return ctx.Err()
	}
	return a.answer(work, ctx, request, history, prompts, updates)
}

var errNoModel = errors.New("no model selected: press Esc, open Settings, then Select model")
