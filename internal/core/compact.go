package core

import (
	"context"
	"errors"

	"jin/internal/provider"
)

// needsCompaction is true once the context fills 80% of the model's window.
// An unknown window (0) never triggers it.
func needsCompaction(size, window int) bool {
	return window > 0 && size*5 >= window*4
}

// compact replaces the whole conversation, system prompt aside, with a
// summary written by the model itself. The summary's output size becomes the
// new context size estimate.
func (a *Agent) compact(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) error {
	if len(*history) < 2 {
		return errors.New("nothing to compact")
	}
	text, usage, err := a.sideRequest(work, ctx, request, *history, a.compactPrompt(), updates)
	if err != nil {
		if usage.Known {
			sendUsage(ctx, updates, request.Model, usage)
		}
		return err
	}
	summary := SummaryMessage(text)
	*history = []provider.Message{(*history)[0], summary}
	a.size = usage.Output
	if !sendHistory(ctx, updates, summary) || !sendCompacted(ctx, updates, request.Model, usage) {
		return ctx.Err()
	}
	return nil
}

// compactIfNeeded is the automatic trigger. A failure is reported but never
// stops the turn; the size resets to 0 so that the check does not retry until
// the next response reports a real one.
func (a *Agent) compactIfNeeded(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) {
	if !needsCompaction(a.size, request.Window) {
		return
	}
	if err := a.compact(work, ctx, request, history, updates); err != nil {
		a.size = 0
		if work.Err() == nil {
			sendUpdate(ctx, updates, UpdateError, "Auto-compaction failed: "+err.Error())
		}
	}
}

func sendCompacted(ctx context.Context, updates chan<- Update, model string, usage provider.Usage) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateCompacted, Model: model, Usage: usage}:
		return true
	}
}
