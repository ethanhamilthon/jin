package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

const bufferError = "exceeded request buffer limit while retrying upstream"

func TestBufferRetryClassification(t *testing.T) {
	for _, tc := range []struct {
		status int
		msg    string
		retry  bool
	}{
		{507, bufferError, true}, {507, "disk full", false}, {507, "", false}, {400, bufferError, false},
	} {
		_, ok := retryDelay(&statusError{status: tc.status, msg: tc.msg}, 1)
		if ok != tc.retry {
			t.Errorf("%d %q: retry = %v", tc.status, tc.msg, ok)
		}
	}
}

func TestBufferRetryLifecycle(t *testing.T) {
	failure := &statusError{status: http.StatusInsufficientStorage, msg: bufferError}
	for _, succeeds := range []bool{true, false} {
		attempts, notices, resets := 0, 0, 0
		_, err := NewClient(Config{}).withRetry(t.Context(), "", func(e StreamEvent) {
			if e.Kind == Notice {
				notices++
			}
			if e.Kind == Reset {
				resets++
			}
		}, func(context.Context) (Response, error) {
			attempts++
			if succeeds && attempts == 2 {
				return Response{}, nil
			}
			return Response{}, failure
		})
		want := maxAttempts
		if succeeds {
			want = 2
		}
		if attempts != want || notices != want-1 || resets != want-1 {
			t.Fatalf("attempts=%d notices=%d resets=%d", attempts, notices, resets)
		}
		if succeeds && err != nil || !succeeds && err != failure {
			t.Fatalf("error = %v", err)
		}
	}
}

func TestBufferRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	attempts := 0
	_, err := NewClient(Config{}).withRetry(ctx, "", func(e StreamEvent) {
		if e.Kind == Notice {
			cancel()
		}
	}, func(context.Context) (Response, error) {
		attempts++
		return Response{}, &statusError{status: 507, msg: bufferError}
	})
	if !errors.Is(err, context.Canceled) || attempts != 1 {
		t.Fatalf("error=%v attempts=%d", err, attempts)
	}
	for attempt := 1; attempt < maxAttempts; attempt++ {
		delay, ok := retryDelay(&statusError{status: 507, msg: bufferError}, attempt)
		if !ok || delay != retryBase*time.Duration(1<<(attempt-1)) {
			t.Fatalf("attempt %d delay=%v", attempt, delay)
		}
	}
}
