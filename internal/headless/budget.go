package headless

import (
	"fmt"
	"strconv"

	"jin/internal/store"
)

// budget is what one run may spend: model requests and dollars, counted
// from the start of this run. A zero limit means no limit.
type budget struct {
	maxCost   float64
	maxTurns  int
	startCost float64
	turns     int
}

func newBudget(opt Options, start store.Usage) *budget {
	return &budget{maxCost: opt.MaxCost, maxTurns: opt.MaxTurns, startCost: start.Cost}
}

// request counts one model request.
func (b *budget) request() { b.turns++ }

// reached names the limit that is used up, or "" while there is room.
func (b *budget) reached(usage store.Usage) string {
	switch {
	case b.maxCost > 0 && usage.Cost-b.startCost >= b.maxCost:
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
