package tools

import "context"

// Adoption is what a bash command needs to carry on as a background task.
type Adoption struct {
	PID, PGID int
	// Command is what was run; Log and Exit are the files of the task. The
	// exit file is written by jin when the process ends.
	Command, Log, Exit string
}

// Background lets the bash tool hand a running command over to the async
// daemon instead of killing it. Adopt returns the task id.
type Background struct {
	Adopt func(Adoption) (id string, err error)
	// Detach asks the running command to move to the background now: it
	// is closed when the user wrote to the agent while a command runs.
	Detach <-chan struct{}
}

type backgroundKey struct{}

// WithBackground tells the bash tool how to hand a command over.
func WithBackground(ctx context.Context, b Background) context.Context {
	return context.WithValue(ctx, backgroundKey{}, b)
}

func backgroundFrom(ctx context.Context) (Background, bool) {
	b, ok := ctx.Value(backgroundKey{}).(Background)
	return b, ok && b.Adopt != nil
}
