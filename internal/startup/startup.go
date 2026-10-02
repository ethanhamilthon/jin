// Package startup builds everything a session needs before its agent can
// run: the system prompt, the compaction and handoff prompts, and the bodies
// of the #prompts. The texts the user wrote for jin (the system prompt file,
// hooks and #prompts) may hold {{commands}}; they are run here, once, when
// the session starts. Text from anywhere else, such as AGENTS.md, is never
// run.
package startup

import (
	"context"
	"sync"

	"jin/internal/core"
	"jin/internal/dyn"
	"jin/internal/hooks"
	"jin/internal/prompts"
	"jin/internal/sysprompt"
)

// Input says what to build.
type Input struct {
	Dir       string
	SessionID string
	ToolNames []string
	// HooksDisabled and PromptsDisabled are the names that are switched off.
	HooksDisabled, PromptsDisabled []string
	// WithPrompts asks for the #prompt bodies; headless runs do not use them.
	WithPrompts bool
	// Env is the environment of the commands; nil means that of jin.
	Env []string
}

// Output is a session's texts, ready to use.
type Output struct {
	System, Compact, Handoff string
	// Prompts maps a #prompt name to its filled-in text.
	Prompts map[string]string
	// Custom is true when ~/.jin/system-prompt.md exists.
	Custom bool
	// Warnings say which commands failed, one line each.
	Warnings []string
}

// Render fills every text and builds the system prompt. All commands of all
// texts run at the same time. onPrompt, when set, is called with the name of
// a #prompt as soon as its commands are done, which lets the screen show
// which prompts are still loading. It may be called from several goroutines.
// When ctx is cancelled the commands stop and their places read
// [command cancelled]; Render still returns a usable Output.
func Render(ctx context.Context, in Input, onPrompt func(name string)) Output {
	opt := dyn.Options{Dir: in.Dir, Env: in.Env}
	sections, _ := sysprompt.Load()
	hookList := hooks.Active(in.HooksDisabled)
	var bodies map[string]string
	if in.WithPrompts {
		bodies = prompts.Bodies(in.PromptsDisabled)
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		warnings []string
	)
	fill := func(text string, after func()) string {
		result := dyn.Expand(ctx, text, opt)
		if len(result.Warnings) > 0 {
			mu.Lock()
			warnings = append(warnings, result.Warnings...)
			mu.Unlock()
		}
		if after != nil {
			after()
		}
		return result.Text
	}

	out := Output{Custom: sections.Custom, Prompts: map[string]string{}}
	hookTexts := make([]string, len(hookList))
	promptTexts := map[string]string{}
	run := func(f func()) {
		wg.Add(1)
		go func() { defer wg.Done(); f() }()
	}
	run(func() { out.System = fill(sections.System, nil) })
	run(func() { out.Compact = fill(sections.Compact, nil) })
	run(func() { out.Handoff = fill(sections.Handoff, nil) })
	for i, hook := range hookList {
		run(func() { hookTexts[i] = fill(hook.Body, nil) })
	}
	for name, body := range bodies {
		run(func() {
			text := fill(body, func() {
				if onPrompt != nil {
					onPrompt(name)
				}
			})
			mu.Lock()
			promptTexts[name] = text
			mu.Unlock()
		})
	}
	wg.Wait()

	out.Prompts = promptTexts
	out.Warnings = sortedUnique(warnings)
	out.System = core.BuildSystemPrompt(core.PromptInput{
		System: out.System, Dir: in.Dir, SessionID: in.SessionID, ToolNames: in.ToolNames, Hooks: hookTexts,
	})
	return out
}

func sortedUnique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
