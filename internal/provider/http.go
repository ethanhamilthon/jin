package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxProviderBody = 4 << 20

func redact(text, key string) string {
	if key != "" {
		text = strings.ReplaceAll(text, key, "[redacted]")
	}
	return text
}

func (c *Client) request(ctx context.Context, method, path string, payload []byte) (int, []byte, error) {
	ctx = c.debugRequest(ctx, path, payload)
	req, cfg, release, err := c.newRequest(ctx, method, path, payload)
	if err != nil {
		return 0, nil, err
	}
	resp, err := completionClient.Do(req)
	if err != nil {
		release()
		c.debugHTTP(ctx, "http_error", map[string]any{"cancelled": ctx.Err() != nil})
		return 0, nil, fmt.Errorf("provider request failed: %s", redact(err.Error(), cfg.APIKey))
	}
	resp.Body = &leasedBody{ReadCloser: resp.Body, release: release}
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

func (c *Client) streamRequest(ctx context.Context, path string, payload []byte) (*http.Response, int, error) {
	ctx = c.debugRequest(ctx, path, payload)
	req, cfg, release, err := c.newRequest(ctx, http.MethodPost, path, payload)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := completionClient.Do(req)
	if err != nil {
		release()
		c.debugHTTP(ctx, "http_error", map[string]any{"cancelled": ctx.Err() != nil})
		if ctx.Err() != nil {
			return nil, 0, err
		}
		return nil, 0, transient(fmt.Errorf("provider request failed: %s", redact(err.Error(), cfg.APIKey)))
	}
	resp.Body = &leasedBody{ReadCloser: resp.Body, release: release}
	c.debugHTTP(ctx, "http_headers", map[string]any{"http_status": resp.StatusCode})
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		return nil, resp.StatusCode, &statusError{
			status:     resp.StatusCode,
			retryAfter: parseRetryAfter(resp.Header.Get("Retry-After"), time.Now()),
			msg:        httpError(path, resp.StatusCode, body, cfg.APIKey).Error(),
		}
	}
	resp.Body = watchBody(ctx, &debugBody{ReadCloser: resp.Body, client: c, ctx: ctx})
	return resp, resp.StatusCode, nil
}
