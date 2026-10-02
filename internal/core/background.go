package core

import (
	"context"

	"jin/internal/tools"
)

// consumedRequest balances Expect: the agent has the message now.
func (a *Agent) consumedRequest(request Request) {
	if request.Interactive {
		a.consumed(1)
	}
}

// backgroundContext lets the bash tool move a command to the background.
func (a *Agent) backgroundContext(ctx context.Context) context.Context {
	if a.background == nil {
		return ctx
	}
	return tools.WithBackground(ctx, tools.Background{Adopt: a.background, Detach: a.toolDetach()})
}
