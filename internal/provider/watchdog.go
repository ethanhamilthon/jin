package provider

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"
)

// DefaultStallTimeout is how long a stream may stay silent before it is
// cancelled and tried again.
const DefaultStallTimeout = 90 * time.Second

var errStalled = errors.New("no data from the provider")

// watchdog cancels an attempt when its response stays silent too long.
// Every read of the response body pushes the deadline back.
type watchdog struct {
	mu       sync.Mutex
	timer    *time.Timer
	limit    time.Duration
	received bool
}

type watchdogKey struct{}

func (w *watchdog) kick() {
	w.mu.Lock()
	w.received = true
	w.timer.Reset(w.limit)
	w.mu.Unlock()
}

// SetStallTimeout changes the silence limit; 0 or less restores the default.
func (c *Client) SetStallTimeout(limit time.Duration) {
	c.mu.Lock()
	c.stall = limit
	c.mu.Unlock()
}

func (c *Client) stallTimeout(effort string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.stall > 0 {
		return c.stall
	}
	return effortTimeout(effort)
}

// watched runs one attempt under a watchdog. A stall comes back as a
// transient error so that withRetry can try once more.
func (c *Client) watched(ctx context.Context, effort string, once func(context.Context) (Response, error)) (Response, error) {
	attempt, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	limit := c.stallTimeout(effort)
	c.Debug("attempt_start", map[string]any{"effort": effort, "silence_timeout_ms": limit.Milliseconds()})
	w := &watchdog{limit: limit}
	w.timer = time.AfterFunc(limit, func() {
		w.mu.Lock()
		phase := "first_byte"
		if w.received {
			phase = "stream"
		}
		w.mu.Unlock()
		c.Debug("stall", map[string]any{"phase": phase, "silence_timeout_ms": limit.Milliseconds()})
		cancel(errStalled)
	})
	defer w.timer.Stop()
	response, err := once(context.WithValue(attempt, watchdogKey{}, w))
	if err != nil && ctx.Err() == nil && errors.Is(context.Cause(attempt), errStalled) {
		return response, &transientError{err: errStalled, stall: true}
	}
	return response, err
}

type watchedBody struct {
	io.ReadCloser
	w *watchdog
}

func (b watchedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.w.kick()
	}
	return n, err
}

// watchBody makes reads of body keep the attempt's watchdog quiet.
func watchBody(ctx context.Context, body io.ReadCloser) io.ReadCloser {
	if w, ok := ctx.Value(watchdogKey{}).(*watchdog); ok {
		return watchedBody{ReadCloser: body, w: w}
	}
	return body
}
