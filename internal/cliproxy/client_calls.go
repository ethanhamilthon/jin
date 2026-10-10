package cliproxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"
)

func (c *Client) Acquire(ctx context.Context) (string, string, func(), error) {
	if err := c.connect(ctx); err != nil {
		return "", "", nil, err
	}
	c.mu.Lock()
	id := c.id
	c.mu.Unlock()
	lease, cancel := context.WithCancel(ctx)
	resp, err := c.request(lease, "GET", "/lease?client="+id, nil)
	if err != nil {
		cancel()
		return "", "", nil, err
	}
	var body struct {
		Endpoint string
		Key      string
	}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		cancel()
		resp.Body.Close()
		return "", "", nil, errors.New("invalid proxy lease")
	}
	var once sync.Once
	return body.Endpoint, body.Key, func() { once.Do(func() { cancel(); resp.Body.Close() }) }, nil
}

func (c *Client) call(ctx context.Context, input command) (json.RawMessage, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	data, _ := json.Marshal(input)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	resp, err := c.request(ctx, "POST", "/command", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}
