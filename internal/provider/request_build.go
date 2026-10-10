package provider

import (
	"bytes"
	"context"
	"errors"
	"net/http"
)

func (c *Client) newRequest(ctx context.Context, method, path string, payload []byte) (*http.Request, Config, func(), error) {
	cfg, release, err := c.prepareRequest(ctx, path, payload)
	if err != nil {
		release()
		return nil, cfg, func() {}, err
	}
	if !cfg.Ready() || cfg.Managed {
		release()
		return nil, cfg, func() {}, errors.New("provider is not configured: open Providers in Settings")
	}
	req, err := http.NewRequestWithContext(ctx, method, cfg.BaseURL+path, bytes.NewReader(payload))
	if err != nil {
		release()
		return nil, cfg, func() {}, errors.New("invalid provider endpoint")
	}
	if cfg.Kind == KindAnthropic {
		req.Header.Set("x-api-key", cfg.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, cfg, release, nil
}
