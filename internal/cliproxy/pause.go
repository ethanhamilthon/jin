package cliproxy

import "context"

func (b *broker) pause(ctx context.Context) (func(), error) {
	b.mu.Lock()
	b.updating = true
	b.mu.Unlock()
	resume := func() { b.mu.Lock(); b.updating = false; b.mu.Unlock(); b.notify() }
	if err := b.waitIdle(ctx); err != nil {
		resume()
		return nil, err
	}
	return resume, nil
}
