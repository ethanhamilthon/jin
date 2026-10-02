package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxProviderBody = 4 << 20

func redact(text, key string) string {
	if key != "" {
		text = strings.ReplaceAll(text, key, "[redacted]")
	}
	return text
}

func (c *Client) newRequest(ctx context.Context, method, path string, payload []byte) (*http.Request, Config, error) {
	cfg := c.Config()
	if !cfg.Ready() {
		return nil, cfg, errors.New("provider is not configured: press Esc, open Settings, then Provider")
	}
	req, err := http.NewRequestWithContext(ctx, method, cfg.BaseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, cfg, errors.New("invalid provider endpoint")
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, cfg, nil
}

func (c *Client) request(ctx context.Context, method, path string, payload []byte) (int, []byte, error) {
	req, cfg, err := c.newRequest(ctx, method, path, payload)
	if err != nil {
		return 0, nil, err
	}
	resp, err := completionClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("provider request failed: %s", redact(err.Error(), cfg.APIKey))
	}
	defer resp.Body.Close()
	limit := int64(maxProviderBody + 1)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limit = 2048
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return 0, nil, errors.New("cannot read provider response")
	}
	if len(body) > maxProviderBody {
		return 0, nil, errors.New("provider response exceeds 4 MiB")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, body, httpError(path, resp.StatusCode, body, cfg.APIKey)
	}
	return resp.StatusCode, body, nil
}

func (c *Client) streamRequest(ctx context.Context, path string, payload []byte) (*http.Response, error) {
	req, cfg, err := c.newRequest(ctx, http.MethodPost, path, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := completionClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("provider request failed: %s", redact(err.Error(), cfg.APIKey))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		return nil, httpError(path, resp.StatusCode, body, cfg.APIKey)
	}
	return resp, nil
}

func httpError(path string, status int, body []byte, key string) error {
	return fmt.Errorf("%s HTTP %d: %s", path, status, redact(strings.TrimSpace(string(body)), key))
}
