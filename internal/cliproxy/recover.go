package cliproxy

import (
	"context"
	"errors"
)

func (b *broker) processDead() bool {
	b.mu.Lock()
	p := b.proxy
	b.mu.Unlock()
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

func (b *broker) restartIfNeeded(ctx context.Context) error {
	if !b.processDead() {
		return nil
	}
	resume, err := b.pause(ctx)
	if err != nil {
		return err
	}
	defer resume()
	p, err := b.startVersion(Installed(b.root).Version)
	if err != nil {
		return errors.New("managed proxy exited and could not restart")
	}
	b.mu.Lock()
	b.proxy = p
	b.mu.Unlock()
	return nil
}

func (b *broker) recoverForLease(ctx context.Context) error {
	if !b.processDead() {
		return nil
	}
	b.operation.Lock()
	defer b.operation.Unlock()
	return b.restartIfNeeded(ctx)
}
