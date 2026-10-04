package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	"jin/internal/provider"
)

// needsCompaction is true once the context fills 80% of the model's window.
// An unknown window (0) never triggers it.
func needsCompaction(size, window int) bool {
	return window > 0 && size*5 >= window*4
}

// compact replaces the whole conversation, system prompt aside, with a
// summary written by the model itself.
func (a *Agent) compact(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update) error {
	return a.compactAs(work, ctx, request, history, updates, "Compacted")
}

// compactAs compacts and labels the divider with what happened: label, and
// the context size before and after. The new size is an estimate of the new
// history plus the tool schemas.
func (a *Agent) compactAs(work, ctx context.Context, request Request, history *[]provider.Message, updates chan<- Update, label string) error {
	before := a.contextSize(*history)
	a.client.Debug("compaction_start", map[string]any{"reason": label, "context_tokens": before, "window": request.Window})
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
	a.reseed(*history)
	a.refreshDue = true
	a.client.Debug("compaction_end", map[string]any{"reason": label, "before": before, "after": a.size})
	if !sendHistory(ctx, updates, summary) || !sendCompacted(ctx, updates, request.Model, usage, compactLabel(label, before, a.size)) {
		return ctx.Err()
	}
	return nil
}

// compactLabel is "Auto-compacted 152K → 3K tokens" when both sizes are known.
func compactLabel(label string, before, after int) string {
	if before <= 0 || after <= 0 {
		return label
	}
	return fmt.Sprintf("%s %s → %s tokens", label, FormatTokens(before), FormatTokens(after))
}

// FormatTokens writes a token count the short way: 950, 3.1K, 1.2M.
func FormatTokens(n int) string {
	switch {
	case n < 1000:
		return strconv.Itoa(n)
	case n < 1_000_000:
		return strconv.FormatFloat(math.Round(float64(n)/100)/10, 'f', -1, 64) + "K"
	}
	return strconv.FormatFloat(math.Round(float64(n)/100_000)/10, 'f', -1, 64) + "M"
}

func sendCompacted(ctx context.Context, updates chan<- Update, model string, usage provider.Usage, label string) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateCompacted, Model: model, Usage: usage, Text: label}:
		return true
	}
}
