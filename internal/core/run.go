package core

import (
	"context"
	"errors"
	"jin/internal/provider"
)

func (a *Agent) Run(ctx context.Context, initial []provider.Message, prompts <-chan Request, updates chan<- Update) {
	defer close(updates)
	history := []provider.Message{{Role: "system", Content: a.systemPrompt}}
	history = append(history, initial...)
	for {
		select {
		case <-ctx.Done():
			return
		case request, ok := <-prompts:
			if !ok {
				return
			}
			a.consumedRequest(request)
			if request.blank() {
				continue
			}
			if !a.turn(ctx, request, &history, prompts, updates) {
				return
			}
		}
	}
}

func (a *Agent) turn(ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) bool {
	if !sendUpdate(ctx, updates, UpdateWorking, "") {
		return false
	}
	turnCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancelTurn = cancel
	a.mu.Unlock()
	err := a.perform(turnCtx, ctx, request, history, prompts, updates)
	a.mu.Lock()
	a.cancelTurn = nil
	a.mu.Unlock()
	interrupted := turnCtx.Err() != nil
	cancel()
	if ctx.Err() != nil {
		return false
	}
	switch {
	case interrupted:
		for _, msg := range InterruptedToolMessages(*history) {
			*history = append(*history, msg)
			sendHistory(ctx, updates, msg)
		}
		sendUpdate(ctx, updates, UpdateInfo, "Interrupted")
	case err != nil:
		sendUpdate(ctx, updates, UpdateError, err.Error())
	}
	return sendDone(ctx, updates, err == nil && !interrupted && request.Kind == RequestPrompt)
}

func (a *Agent) perform(work, ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) error {
	switch request.Kind {
	case RequestCompact:
		return a.compact(work, ctx, request, history, updates)
	case RequestHandoff:
		return a.handoff(work, ctx, request, history, updates)
	}
	a.compactIfNeeded(work, ctx, request, history, updates)
	userMessage := provider.Message{Role: "user", Content: request.Prompt}
	*history = append(*history, userMessage)
	if !sendHistory(ctx, updates, userMessage) {
		return ctx.Err()
	}
	return a.answer(work, ctx, request, history, prompts, updates)
}

var errNoModel = errors.New("no model selected: press Esc, open Settings, then Select model")
