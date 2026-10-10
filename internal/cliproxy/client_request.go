package cliproxy

import (
	"bytes"
	"context"
	"errors"
	"net/http"
)

func (c *Client) request(ctx context.Context, method, path string, data []byte) (*http.Response, error) {
	k, err := loadKeys(c.root)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://managed-proxy"+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+k.Control)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.New("managed CLIProxyAPI is unavailable")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, errors.New("managed CLIProxyAPI operation failed")
	}
	return resp, nil
}
