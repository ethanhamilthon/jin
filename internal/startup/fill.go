package startup

import (
	"context"
	"sync"

	"jin/internal/dyn"
)

// filler runs the commands of many texts at once and collects the warnings.
type filler struct {
	ctx      context.Context
	opt      dyn.Options
	wg       sync.WaitGroup
	mu       sync.Mutex
	warnings []string
}

// fill runs the commands of text and returns it filled in; after, when set,
// is called once it is done.
func (f *filler) fill(text string, after func()) string {
	result := dyn.Expand(f.ctx, text, f.opt)
	if len(result.Warnings) > 0 {
		f.mu.Lock()
		f.warnings = append(f.warnings, result.Warnings...)
		f.mu.Unlock()
	}
	if after != nil {
		after()
	}
	return result.Text
}

// run starts job in the background; wg.Wait waits for all of them.
func (f *filler) run(job func()) {
	f.wg.Add(1)
	go func() { defer f.wg.Done(); job() }()
}
