package headless

import (
	"context"
	"errors"
	"time"

	"jin/internal/core"
	"jin/internal/daemon"
	"jin/internal/store"
)

func runDaemonPrompt(ctx context.Context, db *store.DB, dir string, opt Options, prompt string, out writer) int {
	started := time.Now()
	fail := func(err error) int {
		out.Result(result{Err: err.Error(), Duration: time.Since(started).Milliseconds()})
		return exitError
	}
	if opt.ToolsSet || len(opt.Exclude) > 0 || opt.NoTools || opt.MaxCost > 0 || opt.MaxTurns > 0 {
		return fail(errors.New("tool overrides and per-run budgets are not supported on a shared daemon session; use --no-session for an isolated run"))
	}
	snap, err := daemonPromptSession(ctx, db, dir, opt)
	if err != nil {
		return fail(err)
	}
	id := snap.State.ID
	for !snap.State.Ready {
		select {
		case <-ctx.Done():
			return fail(ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
		snap, err = daemonClient.Snapshot(ctx, id)
		if err != nil {
			return fail(err)
		}
	}
	if opt.Model != "" || opt.Effort != "" || opt.Provider != "" {
		model, effort := snap.State.Model, snap.State.Effort
		if opt.Model != "" {
			model = opt.Model
		}
		if opt.Effort != "" {
			effort = opt.Effort
		}
		if err := daemonClient.Command(ctx, daemon.Command{Action: "model", Session: id, Provider: opt.Provider, Model: model, Effort: effort}, nil); err != nil {
			return fail(err)
		}
	}
	if snap.State.Paused {
		return fail(errors.New("session queue is paused; resume it in a client before sending"))
	}
	before := len(snap.Entries)
	out.Session(sessionInfo{ID: id, Cwd: snap.State.Path, Model: snap.State.Model, Effort: snap.State.Effort, Provider: snap.State.Provider})
	if err := daemonClient.Command(ctx, daemon.Command{Action: "send", Session: id, Text: prompt}, nil); err != nil {
		return fail(err)
	}
	for {
		select {
		case <-ctx.Done():
			return fail(ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
		snap, err = daemonClient.Snapshot(ctx, id)
		if err != nil {
			return fail(err)
		}
		if len(snap.State.Ask) > 0 {
			return fail(errors.New("session is waiting for an answer; answer in TUI or web"))
		}
		if snap.State.Busy {
			continue
		}
		result := result{SessionID: id, Usage: snap.State.Usage, Duration: time.Since(started).Milliseconds()}
		for _, entry := range answerEntries(snap.Entries, before, prompt) {
			result.Text = entry.Text
		}
		for _, entry := range snap.Entries[ourTurnStart(snap.Entries, before, prompt):] {
			if entry.Kind == core.UpdateError {
				result.Err = entry.Text
			}
			if entry.Kind == core.UpdateToolCall {
				out.Progress(entry.Tool + ": " + entry.Text)
			}
		}
		out.Result(result)
		if result.Err != "" {
			return exitError
		}
		return exitOK
	}
}
