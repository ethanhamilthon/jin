package headless

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"jin/internal/provider"
	"jin/internal/store"
)

// budget is what one run may spend: model requests and dollars, counted
// from the start of this run. A zero limit means no limit.
type budget struct {
	maxCost   float64
	maxTurns  int
	startCost float64

	mu      sync.Mutex
	turns   int
	spent   store.Usage
	stopped string
}

var errBudget = errors.New("budget reached")

func newBudget(opt Options, start store.Usage) *budget {
	return &budget{maxCost: opt.MaxCost, maxTurns: opt.MaxTurns, startCost: start.Cost}
}

// gate runs in the agent right before each model request: it adds the
// usage of the response that just completed and refuses the request once a
// limit is used up.
func (r *runState) gate(last provider.Usage) error {
	b := r.budget
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spent.Add(last, r.request.Model, r.table)
	if b.stopped = b.reason(b.spent.Cost); b.stopped != "" {
		return errBudget
	}
	b.turns++
	return nil
}

// compacted counts a compaction, a model request that does not pass the gate.
func (r *runState) compacted(usage provider.Usage, model string) {
	b := r.budget
	b.mu.Lock()
	defer b.mu.Unlock()
	b.turns++
	b.spent.Add(usage, model, r.table)
}

// halted reports whether the gate has refused a request.
func (b *budget) halted() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.stopped != ""
}

// result names the limit that ended the run or that the whole run used up,
// or "" while there is room.
func (b *budget) result(total store.Usage) string {
	if b == nil {
		return ""
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stopped != "" {
		return b.stopped
	}
	return b.reason(total.Cost - b.startCost)
}

func (b *budget) reason(cost float64) string {
	switch {
	case b.maxCost > 0 && cost >= b.maxCost:
		return "max-cost"
	case b.maxTurns > 0 && b.turns >= b.maxTurns:
		return "max-turns"
	}
	return ""
}

// priceError says why --max-cost cannot work: without a price every request
// would look free.
func (r *runState) priceError() error {
	if r.opt.MaxCost <= 0 {
		return nil
	}
	if _, known := r.table.Lookup(r.request.Model); known {
		return nil
	}
	return fmt.Errorf("--max-cost needs a known price, but model %q is not in the price catalogue (limit %s)",
		r.request.Model, strconv.FormatFloat(r.opt.MaxCost, 'f', -1, 64))
}
