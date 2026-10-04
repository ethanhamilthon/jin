package core

import (
	"context"
	"errors"

	"jin/internal/provider"
)

var errAgentStopped = errors.New("agent is not running")

type promptReload struct {
	ctx                      context.Context
	system, compact, handoff string
	refresh                  Refresher
	done                     chan error
}

// ReloadPrompts replaces all rendered prompts at a safe point in the Run loop.
func (a *Agent) ReloadPrompts(ctx context.Context, system, compact, handoff string, refresh Refresher) error {
	request := promptReload{ctx: ctx, system: system, compact: compact, handoff: handoff, refresh: refresh, done: make(chan error, 1)}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-a.runDone:
		return errAgentStopped
	case a.reloads <- request:
	}
	select {
	case err := <-request.done:
		return err
	case <-ctx.Done():
		select {
		case err := <-request.done:
			return err
		default:
			return ctx.Err()
		}
	case <-a.runDone:
		select {
		case err := <-request.done:
			return err
		default:
			return errAgentStopped
		}
	}
}

func (a *Agent) applyReload(runCtx context.Context, request promptReload, history []provider.Message) {
	if err := runCtx.Err(); err != nil {
		request.done <- err
		return
	}
	if err := request.ctx.Err(); err != nil {
		request.done <- err
		return
	}
	a.mu.Lock()
	a.systemPrompt, a.compactText, a.handoffText = request.system, request.compact, request.handoff
	a.refresh = request.refresh
	a.renderedAt, a.refreshDue, a.refreshNote = a.now(), false, ""
	history[0].Content = request.system
	a.mu.Unlock()
	a.reseed(history)
	request.done <- nil
}

func (a *Agent) applyReloads(runCtx context.Context, history []provider.Message) {
	for {
		select {
		case request := <-a.reloads:
			a.applyReload(runCtx, request, history)
		default:
			return
		}
	}
}
