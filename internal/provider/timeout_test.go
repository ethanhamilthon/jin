package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEffortTimeout(t *testing.T) {
	cases := map[string]time.Duration{
		"": 90 * time.Second, "none": 90 * time.Second, "minimal": 90 * time.Second, "weird": 90 * time.Second,
		"low": 120 * time.Second, "medium": 180 * time.Second, " HIGH ": 300 * time.Second,
		"xhigh": 600 * time.Second, "max": 600 * time.Second,
	}
	for effort, want := range cases {
		if got := effortTimeout(effort); got != want {
			t.Errorf("%q: got %v, want %v", effort, got, want)
		}
	}
	client := NewClient(Config{})
	if got := client.stallTimeout("xhigh"); got != 600*time.Second {
		t.Fatalf("policy: %v", got)
	}
	client.SetStallTimeout(5 * time.Second)
	if got := client.stallTimeout("xhigh"); got != 5*time.Second {
		t.Fatalf("override: %v", got)
	}
}

func TestWatchedStallIsTransient(t *testing.T) {
	client := NewClient(Config{})
	client.SetStallTimeout(20 * time.Millisecond)
	_, err := client.watched(t.Context(), "xhigh", func(ctx context.Context) (Response, error) {
		<-ctx.Done()
		return Response{}, ctx.Err()
	})
	var tr *transientError
	if !errors.As(err, &tr) || !tr.stall {
		t.Fatalf("want stall, got %v", err)
	}
}

func TestWatchedCancelIsNotStall(t *testing.T) {
	client := NewClient(Config{})
	client.SetStallTimeout(time.Hour)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := client.watched(ctx, "", func(ctx context.Context) (Response, error) {
		<-ctx.Done()
		return Response{}, ctx.Err()
	})
	var tr *transientError
	if errors.As(err, &tr) || !errors.Is(err, context.Canceled) {
		t.Fatalf("want cancel, got %v", err)
	}
}
