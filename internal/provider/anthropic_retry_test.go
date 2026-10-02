package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAnthropicThinkingRetry400(t *testing.T) {
	tests := []struct {
		name        string
		effort      string
		statusCode  int
		expectRetry bool
		expectErr   bool
	}{
		{
			name:        "retry 400 on thinking and remember disabled",
			effort:      "high",
			statusCode:  http.StatusBadRequest,
			expectRetry: true,
			expectErr:   false,
		},
		{
			name:        "no retry on 500 server error",
			effort:      "high",
			statusCode:  http.StatusInternalServerError,
			expectRetry: false,
			expectErr:   true,
		},
		{
			name:        "no retry on 400 without effort",
			effort:      "",
			statusCode:  http.StatusBadRequest,
			expectRetry: false,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&requestCount, 1)
				data, _ := io.ReadAll(r.Body)
				var body map[string]any
				_ = json.Unmarshal(data, &body)

				if count == 1 && tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte(`{"error":{"message":"thinking not supported"}}`))
					return
				}
				if body["thinking"] != nil {
					t.Errorf("request %d should not have thinking", count)
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("data: {\"type\":\"message_start\",\"message\":{\"role\":\"assistant\"}}\n\n" +
					"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\"}}\n\n" +
					"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"ok\"}}\n\n" +
					"data: {\"type\":\"message_stop\"}\n\n"))
			}))
			defer server.Close()

			client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "k"})
			resp, err := client.Stream(t.Context(), "claude-model", tt.effort, nil, json.RawMessage("[]"), func(StreamEvent) {})
			if tt.expectErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.expectErr {
				if err != nil || resp.Message.Content != "ok" {
					t.Fatalf("stream failed: %+v, err=%v", resp, err)
				}
				if tt.expectRetry && atomic.LoadInt32(&requestCount) != 2 {
					t.Fatalf("expected 2 requests, got %d", requestCount)
				}
				// Verify model remembers disabled thinking on subsequent call
				_, err = client.Stream(t.Context(), "claude-model", tt.effort, nil, json.RawMessage("[]"), func(StreamEvent) {})
				if err != nil {
					t.Fatal(err)
				}
				if tt.expectRetry && atomic.LoadInt32(&requestCount) != 3 {
					t.Fatalf("expected 3 requests total, got %d", requestCount)
				}
			}
		})
	}
}
