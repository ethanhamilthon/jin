package core

import (
	"context"
	"time"

	"jin/internal/provider"
)

const (
	cacheTTL       = 5 * time.Minute
	promptMaxAge   = time.Hour
	refreshedOpen  = "<system-refreshed>"
	refreshedClose = "</system-refreshed>"
	refreshedNote  = refreshedOpen + "instructions were refreshed" + refreshedClose + "\n\n"
)

// Refresher renders the system prompt again, running its commands.
type Refresher func(ctx context.Context) string

// SetRefresher sets how the system prompt is rendered again. Call it before
// Run. Without it the prompt never changes.
func (a *Agent) SetRefresher(refresh Refresher) { a.refresh = refresh }

func (a *Agent) now() time.Time {
	if a.clock != nil {
		return a.clock()
	}
	return time.Now()
}

// systemStale is true when the prompt should be rendered again before the
// next request: after a compaction or handoff, or when the provider cache is
// cold anyway (idle past its TTL) and the prompt is an hour old or from
// another day. A hot cache is never touched.
func (a *Agent) systemStale(now time.Time) bool {
	if a.refresh == nil {
		return false
	}
	if a.refreshDue {
		return true
	}
	idleSince := a.lastDone
	if idleSince.IsZero() {
		idleSince = a.renderedAt
	}
	if now.Sub(idleSince) <= cacheTTL {
		return false
	}
	return now.Sub(a.renderedAt) > promptMaxAge || !sameDay(a.renderedAt, now)
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// refreshSystem renders the prompt again when it is stale and puts it first
// in history. It returns the note for the next user message: empty when the
// text did not change.
func (a *Agent) refreshSystem(ctx context.Context, history []provider.Message) string {
	if !a.systemStale(a.now()) {
		return ""
	}
	text := a.refresh(ctx)
	if ctx.Err() != nil || text == "" {
		return ""
	}
	a.renderedAt, a.refreshDue = a.now(), false
	if text == history[0].Content {
		return ""
	}
	a.systemPrompt, history[0].Content = text, text
	return refreshedNote
}
