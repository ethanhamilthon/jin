package cliproxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

type Client struct {
	root         string
	mu           sync.Mutex
	id           string
	registration io.ReadCloser
	cancel       context.CancelFunc
	http         *http.Client
	closed       bool
}

func NewClient(root string) *Client {
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath(root))
	}}
	return &Client{root: root, http: &http.Client{Transport: transport}}
}

func (c *Client) connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("proxy client is closed")
	}
	if c.id != "" {
		return nil
	}
	if Installed(c.root).Version == "" {
		return errors.New("install CLIProxyAPI in Settings first")
	}
	if err := ensureBroker(ctx, c.root); err != nil {
		return err
	}
	registration, cancel := context.WithCancel(context.Background())
	connected := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-connected:
		}
	}()
	defer close(connected)
	resp, err := c.request(registration, "GET", "/client", nil)
	if err != nil {
		cancel()
		return err
	}
	var body struct{ Client string }
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil || body.Client == "" {
		cancel()
		resp.Body.Close()
		return errors.New("invalid proxy registration")
	}
	c.id, c.registration, c.cancel = body.Client, resp.Body, cancel
	go func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.id == body.Client {
			c.id = ""
			c.registration = nil
			cancel()
		}
	}()
	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.cancel != nil {
		c.cancel()
	}
	if c.registration != nil {
		c.registration.Close()
	}
	c.http.CloseIdleConnections()
}
