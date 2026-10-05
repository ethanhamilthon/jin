package tools

import (
	"context"

	"jin/internal/tasks"
)

// Background gives the bash and task tools the background tasks of the
// session: Owner is whose tasks they are.
type Background struct {
	Tasks *tasks.Manager
	Owner string
	// Detach asks a running bash command to move to the background now: it
	// is closed when the user wrote to the agent while a command runs.
	Detach <-chan struct{}
}

type backgroundKey struct{}

// WithBackground lets the tools start and adopt background tasks.
func WithBackground(ctx context.Context, b Background) context.Context {
	return context.WithValue(ctx, backgroundKey{}, b)
}

func backgroundFrom(ctx context.Context) (Background, bool) {
	b, ok := ctx.Value(backgroundKey{}).(Background)
	return b, ok && b.Tasks != nil
}
