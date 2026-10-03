package headless

import (
	"context"
	"os"
	"time"

	"jin/internal/core"
	"jin/internal/store"
)

// outcome is what the update loop saw by the time the request ended.
type outcome struct {
	answer, runErr               string
	final, interrupted, timedOut bool
}

// run sends the request and waits for it, printing progress, until it ends,
// times out, or a signal interrupts it.
func (r *runState) run(ctx context.Context, signals <-chan os.Signal) int {
	r.table = <-r.prices
	entry, known := r.table.Lookup(r.request.Model)
	r.request.Window, r.request.NoVision = entry.MaxInputTokens, known && entry.VisionKnown && !entry.Vision
	r.out.Session(r.id, r.dir, r.request.Model, r.request.Effort)

	runCtx, stop := context.WithCancel(ctx)
	defer stop()
	requests := make(chan core.Request, 1)
	updates := make(chan core.Update, 64)
	go r.agent.Run(runCtx, core.SinceLastSummary(r.history), requests, updates)
	requests <- r.request
	started := time.Now()
	var timeout <-chan time.Time
	if r.opt.Timeout > 0 {
		timer := time.NewTimer(r.opt.Timeout)
		defer timer.Stop()
		timeout = timer.C
	}
	usage := r.record.Usage
	var o outcome
	retry := time.NewTicker(50 * time.Millisecond)
	defer retry.Stop()
	for finished := false; !finished; {
		select {
		case u, ok := <-updates:
			finished = !ok || r.apply(u, &o, &usage)
		case <-timeout:
			o.timedOut, timeout = true, nil
			r.agent.Interrupt()
		case <-signals:
			o.interrupted = true
			r.agent.Interrupt()
		case <-retry.C:
			// An interrupt that lands before the turn starts is lost, so
			// keep asking until the run ends.
			if o.timedOut || o.interrupted {
				r.agent.Interrupt()
			}
		}
	}
	close(requests)
	return r.finish(o, result{Text: o.answer, SessionID: r.id, Duration: time.Since(started).Milliseconds(), Usage: usage})
}

// apply handles one update; it reports whether the request has ended.
func (r *runState) apply(u core.Update, o *outcome, usage *store.Usage) bool {
	switch u.Kind {
	case core.UpdateHistory:
		if r.save {
			if err := r.db.AppendMessage(r.id, u.Message); err != nil {
				r.out.Progress("jin: history was not saved: " + err.Error())
			}
		}
		r.out.Message(u.Message)
		if u.Message.Role == "assistant" && len(u.Message.ToolCalls) == 0 {
			o.answer = u.Message.Content
		}
	case core.UpdateUsage, core.UpdateCompacted:
		usage.Add(u.Usage, u.Model, r.table)
		if u.Kind == core.UpdateCompacted && u.Usage.Known {
			usage.Context = u.Usage.Output
		}
		if u.Kind == core.UpdateCompacted && u.Text != "" {
			r.out.Progress(u.Text)
		}
		if r.save {
			_ = r.db.SaveUsage(r.id, *usage)
		}
	case core.UpdateToolCall:
		r.out.Progress(u.Tool + ": " + u.Text)
	case core.UpdateInfo:
		r.out.Progress(u.Text)
	case core.UpdateError:
		o.runErr = u.Text
		r.out.Progress("error: " + u.Text)
	case core.UpdateDone:
		o.final = u.Final
		return true
	}
	return false
}

func (r *runState) finish(o outcome, res result) int {
	code := exitOK
	switch {
	case o.interrupted:
		res.Err, code = "interrupted", exitInterrupted
	case o.timedOut:
		res.Err, code = "timed out after "+r.opt.Timeout.String(), exitError
	case o.runErr != "":
		res.Err, code = o.runErr, exitError
	case !o.final:
		res.Err, code = "the request did not finish", exitError
	}
	r.out.Result(res)
	return code
}
