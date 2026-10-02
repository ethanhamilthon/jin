// Package dyn fills prompts with live data. A {{command}} in a text is run
// with bash and replaced by its output; \{{ keeps a literal {{. It is meant
// for text the user wrote and trusts: the system prompt file, hooks and
// #prompts. It must never see text from a repository or a task output.
package dyn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
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

// Result is a filled text and what went wrong on the way.
type Result struct {
	Text     string
	Warnings []string
}

type piece struct {
	literal string
	command string
	isCmd   bool
}

// split cuts a text into literal pieces and commands. An empty {{}} and a {{
// that is never closed stay literal text.
func split(text string) []piece {
	var pieces []piece
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			pieces = append(pieces, piece{literal: lit.String()})
			lit.Reset()
		}
	}
	for i := 0; i < len(text); {
		switch {
		case strings.HasPrefix(text[i:], `\{{`):
			lit.WriteString("{{")
			i += 3
		case strings.HasPrefix(text[i:], "{{"):
			end := strings.Index(text[i+2:], "}}")
			if end < 0 {
				lit.WriteString(text[i:])
				i = len(text)
				break
			}
			command := strings.TrimSpace(text[i+2 : i+2+end])
			if command == "" {
				lit.WriteString(text[i : i+2+end+2])
			} else {
				flush()
				pieces = append(pieces, piece{command: command, isCmd: true})
			}
			i += 2 + end + 2
		default:
			lit.WriteByte(text[i])
			i++
		}
	}
	flush()
	return pieces
}

// Has reports whether a text holds a command, so callers can skip the work.
func Has(text string) bool {
	for _, p := range split(text) {
		if p.isCmd {
			return true
		}
	}
	return false
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
	for i, p := range pieces {
		if p.isCmd {
			b.WriteString(outputs[i])
		} else {
			b.WriteString(p.literal)
		}
	}
	return Result{Text: b.String(), Warnings: sortStable(warnings)}
}

func marker(err error) string {
	if errors.Is(err, context.Canceled) {
		return "[command cancelled]"
	}
	return "[command failed: " + err.Error() + "]"
}

// run starts one command in its own process group and waits for it, within
// the timeout. Trailing newlines are cut off its output.
func run(ctx context.Context, command string, opt Options) (string, error) {
	select {
	case slots <- struct{}{}:
		defer func() { <-slots }()
	case <-ctx.Done():
		return "", context.Canceled
	}
	runCtx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = opt.Dir
	cmd.Env = opt.Env
	setGroup(cmd)
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	select {
	case err := <-finished:
		if err != nil {
			return "", failure(err, &stderr)
		}
		return strings.TrimRight(out.String(), "\r\n"), nil
	case <-runCtx.Done():
		killGroup(cmd)
		<-finished
		if ctx.Err() != nil {
			return "", context.Canceled
		}
		return "", fmt.Errorf("timed out after %s", Timeout)
	}
}

func failure(err error, stderr *bytes.Buffer) error {
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		return err
	}
	msg := strings.TrimSpace(stderr.String())
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	if msg == "" {
		return fmt.Errorf("exit status %d", exit.ExitCode())
	}
	return fmt.Errorf("exit status %d: %s", exit.ExitCode(), msg)
}

// sortStable keeps warnings in a steady order whatever finished first.
func sortStable(w []string) []string {
	for i := 1; i < len(w); i++ {
		for j := i; j > 0 && w[j] < w[j-1]; j-- {
			w[j], w[j-1] = w[j-1], w[j]
		}
	}
	return w
}
