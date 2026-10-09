// Package dyn fills prompts with live data. A {{command}} in a text is run
// with bash and replaced by its output; \{{ keeps a literal {{. It is meant
// for text the user wrote and trusts: the system prompt file, hooks and
// #prompts. It must never see text from a repository or a task output.
package dyn

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Timeout is how long one command may run.
var Timeout = 10 * time.Second

// maxParallel limits the commands that run at once, across all calls.
const maxParallel = 8

var slots = make(chan struct{}, maxParallel)

// Options says where commands run.
type Options struct {
	Dir string
	// Env is the environment of the commands; nil means the environment of jin.
	Env []string
}

// Segment is one piece of a filled text: literal text (Command empty) or the
// output of one command.
type Segment struct{ Command, Text string }

// Result is a filled text, the pieces it is made of, and what went wrong on
// the way.
type Result struct {
	Text     string
	Segments []Segment
	Warnings []string
}

// Expand runs the commands of a text, in parallel, and puts their output in
// their place. A command that fails or times out leaves a marker in its place
// and a warning; the rest of the text is still filled in. A cancelled context
// stops the commands and leaves [command cancelled].
func Expand(ctx context.Context, text string, opt Options) Result {
	pieces := split(text)
	outputs := make([]string, len(pieces))
	var mu sync.Mutex
	var warnings []string
	var wg sync.WaitGroup
	for i, p := range pieces {
		if !p.isCmd {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := run(ctx, p.command, opt)
			if err != nil {
				// A cancelled command is what the user asked for, not a failure.
				if !errors.Is(err, context.Canceled) {
					mu.Lock()
					warnings = append(warnings, fmt.Sprintf("{{%s}}: %v", p.command, err))
					mu.Unlock()
				}
				out = marker(err)
			}
			outputs[i] = out
		}()
	}
	wg.Wait()
	var b strings.Builder
	segments := make([]Segment, len(pieces))
	for i, p := range pieces {
		if p.isCmd {
			segments[i] = Segment{Command: p.command, Text: outputs[i]}
		} else {
			segments[i] = Segment{Text: p.literal}
		}
		b.WriteString(segments[i].Text)
	}
	return Result{Text: b.String(), Segments: segments, Warnings: sortStable(warnings)}
}
