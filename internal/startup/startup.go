// Package startup builds the built-in service prompts and the bodies of the
// #prompts. The system prompt is built in code and runs no command. Hooks and
// #prompts may hold {{commands}}; they run once when the session starts. Text
// from anywhere else, such as AGENTS.md, is never run.
package startup

import (
	"context"

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
	// PromptsDisabled are the names of the #prompts that are switched off.
	PromptsDisabled []string
	// WithPrompts asks for the #prompt bodies; headless runs do not use them.
	WithPrompts bool
	// Env is the environment of the commands; nil means that of jin.
	Env []string
	// Hooks returns the hook switches when the system prompt is built; nil means no hooks.
	Hooks func() hooks.Settings
}

// Output is a session's texts, ready to use.
type Output struct {
	System, Compact, Handoff string
	// Prompts maps a #prompt name to its filled-in text.
	Prompts map[string]string
	// Custom remains false for source compatibility. Built-in prompts are always used.
	Custom bool
	// Warnings say which commands failed, one line each.
	Warnings []string
}

// Render fills every text and builds the system prompt. All commands of all
// texts run at the same time. onPrompt, when set, is called with the name of
// a #prompt as soon as its commands are done, which lets the screen show
// which prompts are still loading. It may be called from several goroutines.
// The system prompt is built in code. When ctx is cancelled, hook and prompt commands stop and return
// [command cancelled]; Render still returns a usable Output.
func Render(ctx context.Context, in Input, onPrompt func(name string)) Output {
	env := commandEnv(in)
	opt := dyn.Options{Dir: in.Dir, Env: env}
	sections, _ := sysprompt.Load()
	var loadWarnings []string
	var bodies map[string]string
	if in.WithPrompts {
		var promptWarnings []string
		bodies, promptWarnings = prompts.LoadBodies(in.PromptsDisabled)
		loadWarnings = append(loadWarnings, promptWarnings...)
	}

	f := &filler{ctx: ctx, opt: opt, warnings: loadWarnings}
	fill, run := f.fill, f.run

	out := Output{Custom: false, Prompts: map[string]string{}}
	var system []core.PromptPart
	promptTexts := map[string]string{}
	run(func() {
		var warnings []string
		system, warnings = systemParts(ctx, in, sections.System, env)
		f.mu.Lock()
		f.warnings = append(f.warnings, warnings...)
		f.mu.Unlock()
	})
	run(func() { out.Compact = fill(sections.Compact, nil) })
	run(func() { out.Handoff = fill(sections.Handoff, nil) })
	for name, body := range bodies {
		run(func() {
			text := fill(body, func() {
				if onPrompt != nil {
					onPrompt(name)
				}
			})
			f.mu.Lock()
			promptTexts[name] = text
			f.mu.Unlock()
		})
	}
	f.wg.Wait()

	out.Prompts = promptTexts
	out.Warnings = sortedUnique(f.warnings)
	out.System = core.BuildSystemPrompt(system)
	return out
}
