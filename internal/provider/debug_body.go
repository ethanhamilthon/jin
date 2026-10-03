package provider

import (
	"context"
	"io"
	"sync"
)

type debugBody struct {
	io.ReadCloser
	client *Client
	ctx    context.Context
	first  sync.Once
}

func (b *debugBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.first.Do(func() { b.client.debugHTTP(b.ctx, "first_byte", map[string]any{}) })
	}
	return n, err
}
