package agentkit

import (
	"context"

	"jin/internal/core"
	"jin/internal/startup"
)

// Apply gives the agent the rendered system prompt and the compaction and
// handoff instructions. Call it before the agent runs.
func Apply(agent *core.Agent, out startup.Output) {
	agent.SetSystemPrompt(out.System)
	agent.SetSidePrompts(out.Compact, out.Handoff)
}

// Refresher renders the system prompt again for the given input, without the
// #prompt bodies.
func Refresher(in startup.Input) core.Refresher {
	in.WithPrompts = false
	return func(ctx context.Context) string { return startup.Render(ctx, in, nil).System }
}
