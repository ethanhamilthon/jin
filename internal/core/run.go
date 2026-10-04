package core

import (
	"context"
	"jin/internal/provider"
)

func (a *Agent) Run(ctx context.Context, initial []provider.Message, prompts <-chan Request, updates chan<- Update) {
	defer func() {
		close(updates)
		a.runDoneOnce.Do(func() { close(a.runDone) })
	}()
	a.mu.Lock()
	history := []provider.Message{{Role: "system", Content: a.systemPrompt}}
	a.mu.Unlock()
	history = append(history, initial...)
	a.startSize(history)
	for {
		if ctx.Err() != nil {
			return
		}
		a.applyReloads(ctx, history)
		select {
		case <-ctx.Done():
			return
		case reload := <-a.reloads:
			a.applyReload(ctx, reload, history)
		case request, ok := <-prompts:
			if !ok {
				return
			}
			select {
			case reload := <-a.reloads:
				a.applyReload(ctx, reload, history)
			default:
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
	a.lastDone = a.now()
	if err == nil && request.Kind != RequestPrompt {
		a.refreshDue = true
	}
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
