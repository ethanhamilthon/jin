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

// backgroundContext gives the bash and task tools the agent's background tasks.
func (a *Agent) backgroundContext(ctx context.Context) context.Context {
	if a.tasks == nil {
		return ctx
	}
	return tools.WithBackground(ctx, tools.Background{Tasks: a.tasks, Owner: a.tasksOwner, Detach: a.toolDetach()})
}
