// Package startup builds everything a session needs before its agent can
// run: the system prompt, the compaction and handoff prompts, and the bodies
// of the #prompts. The texts the user wrote for jin (the system prompt file,
// hooks and #prompts) may hold {{commands}}; they are run here, once, when
// the session starts. Text from anywhere else, such as AGENTS.md, is never
// run.
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
	ToolNames []string
	// HooksDisabled and PromptsDisabled are the names that are switched off.
	HooksDisabled, PromptsDisabled []string
	// ProjectHooks lets the hooks in Dir/.jin/hooks run; the user trusted them.
	ProjectHooks bool
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
	hookList := hooks.ActiveIn(in.Dir, in.HooksDisabled, in.ProjectHooks)
	var bodies map[string]string
	if in.WithPrompts {
		bodies = prompts.Bodies(in.PromptsDisabled)
	}

	f := &filler{ctx: ctx, opt: opt}
	fill, run := f.fill, f.run

	out := Output{Custom: sections.Custom, Prompts: map[string]string{}}
	hookTexts := make([]string, len(hookList))
	promptTexts := map[string]string{}
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
			f.mu.Lock()
			promptTexts[name] = text
			f.mu.Unlock()
		})
	}
	f.wg.Wait()

	out.Prompts = promptTexts
	out.Warnings = sortedUnique(f.warnings)
	out.System = core.BuildSystemPrompt(core.PromptInput{
		System: out.System, Dir: in.Dir, SessionID: in.SessionID, ToolNames: in.ToolNames, Hooks: hookTexts,
	})
	return out
}
