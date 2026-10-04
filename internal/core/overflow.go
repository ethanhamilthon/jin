package core

import (
	"context"
	"fmt"

	"jin/internal/provider"
)

// compactIfNeeded is the automatic trigger. From 70% of the window it
// prunes old tool output; it compacts only when the context is still at 80%.
// A failure is reported but never stops the turn; the size resets so that
// the check does not retry until the next response reports a real one.
func (a *Agent) compactIfNeeded(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) {
	size := a.contextSize(*history)
	if needsPrune(size, request.Window) && a.prune(history, pruneKeepTurns) {
		a.client.Debug("prune", map[string]any{"before": size, "after": a.contextSize(*history), "window": request.Window})
		size = a.contextSize(*history)
	}
	if !needsCompaction(size, request.Window) {
		return
	}
	sendUpdate(ctx, updates, UpdateInfo, fmt.Sprintf("Context is %d%% full, compacting the conversation...", size*100/request.Window))
	if err := a.compactAs(work, ctx, request, history, updates, "Auto-compacted"); err != nil {
		a.client.Debug("compaction_failed", map[string]any{"context_tokens": size, "window": request.Window})
		a.size, a.mark = 0, len(*history)
		if work.Err() == nil {
			sendUpdate(ctx, updates, UpdateError, "Auto-compaction failed: "+err.Error())
		}
	}
}

// recoverOverflow makes room after the provider refused the request as too
// long: it prunes old tool output, then compacts unless the estimate shows
// the prune was enough. The caller retries the request once.
func (a *Agent) recoverOverflow(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) error {
	before := a.contextSize(*history)
	a.client.Debug("context_overflow", map[string]any{"context_tokens": before, "window": request.Window})
	sendUpdate(ctx, updates, UpdateInfo, "The context window is full: freeing space and retrying...")
	pruned := a.prune(history, pruneKeepTurns)
	if pruned && request.Window > 0 && before >= request.Window && !needsCompaction(a.contextSize(*history), request.Window) {
		return nil
	}
	if err := a.compactAs(work, ctx, request, history, updates, "Compacted after overflow"); err != nil {
		return fmt.Errorf("the conversation does not fit into the context window, and compaction failed: %w", err)
	}
	return nil
}
