package provider

import (
	"net/http"
	"time"
)

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
