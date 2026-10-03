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

// statusError is a non-2xx answer from the provider.
type statusError struct {
	status     int
	retryAfter time.Duration
	msg        string
}

func (e *statusError) Error() string { return e.msg }

// transientError is a failure worth retrying: a dropped connection, a stream
// cut short, an overloaded provider or a stalled stream.
type transientError struct {
	err   error
	stall bool
}

func (e *transientError) Error() string { return e.err.Error() }
func (e *transientError) Unwrap() error { return e.err }

func transient(err error) error { return &transientError{err: err} }

func retryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout, 529:
		return true
	}
	return false
}

// retryDelay says whether err is worth another attempt and how long to wait.
func retryDelay(err error, attempt int) (time.Duration, bool) {
	backoff := retryBase << (attempt - 1)
	var status *statusError
	if errors.As(err, &status) {
		if !retryableStatus(status.status) {
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
func (c *Client) withRetry(ctx context.Context, onEvent func(StreamEvent), once func(context.Context) (Response, error)) (Response, error) {
	stalls := 0
	for attempt := 1; ; attempt++ {
		response, err := c.watched(ctx, once)
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
		onEvent(StreamEvent{Kind: Notice, Text: retryNotice(err, delay, attempt)})
		select {
		case <-ctx.Done():
			return Response{}, ctx.Err()
		case <-time.After(delay):
		}
	}
}
