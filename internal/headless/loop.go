package headless

import (
	"context"
	"time"

	"jin/internal/core"
)

// run sends the request and waits for it, printing progress, until it ends,
// times out, or a signal interrupts it.
func (r *runState) run(ctx context.Context) int {
	r.table = <-r.prices
	entry, known := r.table.Lookup(r.request.Model)
	r.request.Window, r.request.NoVision = entry.MaxInputTokens, known && entry.VisionKnown && !entry.Vision
	r.out.Session(r.id, r.dir, r.request.Model, r.request.Effort)

	runCtx, stop := context.WithCancel(context.WithoutCancel(ctx))
	defer stop()
	requests := make(chan core.Request, 1)
	updates := make(chan core.Update, 64)
	go r.agent.Run(runCtx, core.SinceLastSummary(r.history), requests, updates)
	requests <- r.request
	started := time.Now()
	limit := ctx.Done()
	usage := r.record.Usage
	var o outcome
	retry := time.NewTicker(50 * time.Millisecond)
	defer retry.Stop()
	for finished := false; !finished; {
		select {
		case u, ok := <-updates:
			finished = !ok || r.apply(u, &o, &usage)
		case <-limit:
			o.timedOut = timedOut(ctx)
			o.interrupted, limit = !o.timedOut, nil
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
