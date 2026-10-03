package core

import (
	"jin/internal/tools"
	"time"
)

// SetBackground tells the agent how to hand a bash command over to the async
// daemon. Without it a command whose time is up is killed.
func (a *Agent) SetBackground(adopt func(tools.Adoption) (string, error)) {
	a.background = adopt
}

// Expect counts a user message that is on its way to the agent. Call it
// before the message is queued; a tool that starts or is running meanwhile
// moves to the background instead of making the user wait.
func (a *Agent) Expect() {
	if a != nil {
		a.waiting.Add(1)
	}
}

func (a *Agent) consumed(n int32) {
	if a.waiting.Add(-n) < 0 {
		a.waiting.Store(0)
	}
}

// DetachTools asks the running tool calls to move to the background now. It
// is called after a user message was queued, so a long bash command does not
// hold the message back.
func (a *Agent) DetachTools() {
	if a == nil {
		return
	}
	a.detachMu.Lock()
	defer a.detachMu.Unlock()
	if a.detach != nil {
		close(a.detach)
		a.detach = nil
	}
}

// toolDetach is the channel a tool call watches: it closes when a user
// message is waiting. A message that was already waiting when the tool
// starts closes it after a short moment, so quick tools still finish.
func (a *Agent) toolDetach() <-chan struct{} {
	a.detachMu.Lock()
	defer a.detachMu.Unlock()
	if a.detach == nil {
		a.detach = make(chan struct{})
	}
	ch := a.detach
	if a.waiting.Load() > 0 {
		go func() {
			time.Sleep(detachDelay)
			a.DetachTools()
		}()
	}
	return ch
}

// detachDelay is how long a tool may run while a user message already waits.
const detachDelay = time.Second
