package provider

import (
	"context"
	"io"
	"sync"
)

type RequestTarget func(context.Context, Config, string, []byte) (Config, func(), error)

func (c *Client) SetTarget(target RequestTarget, modelPrefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.target, c.modelPrefix = target, modelPrefix
}

func (c *Client) prepareRequest(ctx context.Context, path string, payload []byte) (Config, func(), error) {
	c.mu.RLock()
	cfg, target := c.cfg, c.target
	c.mu.RUnlock()
	if target == nil {
		return cfg, func() {}, nil
	}
	resolved, release, err := target(ctx, cfg, path, payload)
	if release == nil {
		release = func() {}
	}
	return resolved, release, err
}

type leasedBody struct {
	io.ReadCloser
	once    sync.Once
	release func()
}

func (b *leasedBody) Close() error {
	defer b.once.Do(b.release)
	return b.ReadCloser.Close()
}
