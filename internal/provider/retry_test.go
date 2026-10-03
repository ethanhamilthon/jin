package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const okStream = "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n"

func streamOnce(t *testing.T, client *Client) (Response, []string, error) {
	t.Helper()
	var notices []string
	response, err := client.Stream(t.Context(), "m", "", nil, json.RawMessage("[]"), func(e StreamEvent) {
		if e.Kind == Notice {
			notices = append(notices, e.Text)
		}
	})
	return response, notices, err
}

func TestStreamRetriesTransientStatus(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(okStream))
	}))
	defer server.Close()
	response, notices, err := streamOnce(t, NewClient(Config{BaseURL: server.URL, APIKey: "k"}))
	if err != nil || response.Message.Content != "ok" || count != 2 {
		t.Fatalf("retry: %+v %v count=%d", response, err, count)
	}
	if len(notices) != 1 || !strings.Contains(notices[0], "HTTP 503") || !strings.Contains(notices[0], "(2/5)") {
		t.Fatalf("notices: %q", notices)
	}
}

func TestStreamDoesNotRetryClientError(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	if _, _, err := streamOnce(t, NewClient(Config{BaseURL: server.URL, APIKey: "k"})); err == nil || count != 1 {
		t.Fatalf("401 must fail at once: %v count=%d", err, count)
	}
}

func TestStreamRetriesStallOnce(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) == 1 {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return
		}
		_, _ = w.Write([]byte(okStream))
	}))
	defer server.Close()
	client := NewClient(Config{BaseURL: server.URL, APIKey: "k"})
	client.SetStallTimeout(50 * time.Millisecond)
	response, notices, err := streamOnce(t, client)
	if err != nil || response.Message.Content != "ok" {
		t.Fatalf("stall retry: %+v %v", response, err)
	}
	if len(notices) != 1 || !strings.Contains(notices[0], "Stream stalled") {
		t.Fatalf("notices: %q", notices)
	}
}

func TestRetryDelay(t *testing.T) {
	if d, ok := retryDelay(&statusError{status: 429, retryAfter: 7 * time.Second}, 1); !ok || d != 7*time.Second {
		t.Fatalf("Retry-After: %v %v", d, ok)
	}
	if d, ok := retryDelay(transient(errStalled), 3); !ok || d != 4*retryBase {
		t.Fatalf("backoff: %v %v", d, ok)
	}
	if _, ok := retryDelay(&statusError{status: 400}, 1); ok {
		t.Fatal("400 is not transient")
	}
	now := time.Now()
	if d := parseRetryAfter(now.Add(10*time.Second).UTC().Format(http.TimeFormat), now); d < 9*time.Second || d > 10*time.Second {
		t.Fatalf("http date: %v", d)
	}
}
