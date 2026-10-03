package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicSSEParsing(t *testing.T) {
	tests := []struct {
		name       string
		sse        string
		wantText   string
		wantThink  string
		wantUsage  Usage
		wantEvents int
		checkCalls func(t *testing.T, calls []ToolCall)
	}{
		{
			name: "text and thinking deltas",
			sse: "data: {\"type\":\"message_start\",\"message\":{\"role\":\"assistant\",\"usage\":{\"input_tokens\":10,\"output_tokens\":1,\"cache_read_input_tokens\":5}}}\n\n" +
				"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"thinking\"}}\n\n" +
				"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"thinking_delta\",\"thinking\":\"think-fast\"}}\n\n" +
				"data: {\"type\":\"content_block_stop\",\"index\":0}\n\n" +
				"data: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"text\"}}\n\n" +
				"data: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"text_delta\",\"text\":\"hello world\"}}\n\n" +
				"data: {\"type\":\"content_block_stop\",\"index\":1}\n\n" +
				"data: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":20}}\n\n" +
				"data: {\"type\":\"message_stop\"}\n\n",
			wantText:   "hello world",
			wantThink:  "think-fast",
			wantUsage:  Usage{Input: 15, Output: 20, CachedInput: 5, CacheKnown: true, Known: true},
			wantEvents: 2,
		},
		{
			name: "tool use with input json deltas",
			sse: "data: {\"type\":\"message_start\",\"message\":{\"role\":\"assistant\",\"usage\":{\"input_tokens\":5}}}\n\n" +
				"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"call_1\",\"name\":\"read\"}}\n\n" +
				"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"path\\\": \"}}\n\n" +
				"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"\\\"foo.go\\\"}\"}}\n\n" +
				"data: {\"type\":\"content_block_stop\",\"index\":0}\n\n" +
				"data: {\"type\":\"message_stop\"}\n\n",
			wantUsage: Usage{Input: 5, Known: true},
			checkCalls: func(t *testing.T, calls []ToolCall) {
				if len(calls) != 1 || calls[0].ID != "call_1" || calls[0].Function.Name != "read" {
					t.Fatalf("unexpected calls: %+v", calls)
				}
				if calls[0].Function.Arguments != `{"path": "foo.go"}` {
					t.Fatalf("bad arguments: %q", calls[0].Function.Arguments)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte(tt.sse))
			}))
			defer server.Close()

			client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "k"})
			var events []StreamEvent
			resp, err := client.Stream(t.Context(), "claude", "", nil, json.RawMessage("[]"), func(e StreamEvent) {
				events = append(events, e)
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Message.Content != tt.wantText || resp.Message.ReasoningContent != tt.wantThink {
				t.Fatalf("content/reasoning mismatch: got text=%q, think=%q", resp.Message.Content, resp.Message.ReasoningContent)
			}
			if resp.Usage != tt.wantUsage {
				t.Fatalf("usage mismatch: got %+v, want %+v", resp.Usage, tt.wantUsage)
			}
			if tt.wantEvents > 0 && len(events) != tt.wantEvents {
				t.Fatalf("events count mismatch: got %d, want %d", len(events), tt.wantEvents)
			}
			if tt.checkCalls != nil {
				tt.checkCalls(t, resp.Message.ToolCalls)
			}
		})
	}
}
