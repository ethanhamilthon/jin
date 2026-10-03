package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAnthropicCacheControlFallback(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(data, &body)
		if calls.Add(1) == 1 {
			if body["cache_control"] == nil {
				t.Error("first request must ask for caching")
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"Top-level cache_control has ttl='5m' but the target block already has cache_control"}}`))
			return
		}
		if body["cache_control"] != nil || body["thinking"] == nil {
			t.Errorf("retry must drop only cache_control: %s", data)
		}
		_, _ = w.Write([]byte(`data: {"type":"message_start","message":{"usage":{"input_tokens":1}}}` + "\n\n" +
			`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "k"})
	if _, err := client.Stream(t.Context(), "c", "high", nil, nil, func(StreamEvent) {}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || client.modelSupportsCache("c") || !client.modelSupportsThinking("c") {
		t.Fatalf("calls=%d", calls.Load())
	}
}
