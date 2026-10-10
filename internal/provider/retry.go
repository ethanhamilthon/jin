package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxAttempts   = 5
	maxRetryDelay = 60 * time.Second
)

// retryBase is the first backoff step; it doubles on each attempt.
var retryBase = time.Second

// retryDelay says whether err is worth another attempt and how long to wait.
func retryDelay(err error, attempt int) (time.Duration, bool) {
	backoff := retryBase << (attempt - 1)
	var status *statusError
	if errors.As(err, &status) {
		if !retryableStatus(status.status) && !(status.status == http.StatusInsufficientStorage && strings.Contains(status.msg, "exceeded request buffer limit while retrying upstream")) {
			return 0, false
		}
		if status.retryAfter > 0 {
			backoff = status.retryAfter
		}
		return min(backoff, maxRetryDelay), true
	}
	var tr *transientError
	if errors.As(err, &tr) {
		return min(backoff, maxRetryDelay), true
	}
	return 0, false
}

func parseRetryAfter(header string, now time.Time) time.Duration {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if at, err := http.ParseTime(header); err == nil && at.After(now) {
		return at.Sub(now)
	}
	return 0
}

func retryNotice(err error, delay time.Duration, attempt int) string {
	reason := "Provider error"
	var status *statusError
	var tr *transientError
	switch {
	case errors.As(err, &status):
		reason = fmt.Sprintf("Provider returned HTTP %d", status.status)
	case errors.As(err, &tr) && tr.stall:
		reason = "Stream stalled"
	case errors.As(err, &tr):
		reason = "Connection failed: " + tr.err.Error()
	}
	return fmt.Sprintf("%s, retrying in %s (%d/%d)", reason, delay.Round(time.Second), attempt+1, maxAttempts)
}

// withRetry runs one streaming attempt at a time until it succeeds, fails
// for good, or runs out of attempts. A stalled stream is retried only once.
func (c *Client) withRetry(ctx context.Context, effort string, onEvent func(StreamEvent), once func(context.Context) (Response, error)) (Response, error) {
	stalls := 0
	for attempt := 1; ; attempt++ {
		response, err := c.watched(ctx, effort, once)
		if err == nil || ctx.Err() != nil || attempt == maxAttempts {
			return response, err
		}
		var tr *transientError
		if errors.As(err, &tr) && tr.stall {
			if stalls++; stalls > 1 {
				return response, err
			}
		}
		delay, ok := retryDelay(err, attempt)
		if !ok {
			return response, err
		}
		c.Debug("retry", map[string]any{"attempt": attempt, "delay_ms": delay.Milliseconds()})
		onEvent(StreamEvent{Kind: Reset, Usage: response.Usage})
		onEvent(StreamEvent{Kind: Notice, Text: retryNotice(err, delay, attempt)})
		select {
		case <-ctx.Done():
			return Response{}, ctx.Err()
		case <-time.After(delay):
		}
	}
}
