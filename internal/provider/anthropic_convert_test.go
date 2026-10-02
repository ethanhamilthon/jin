package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicMessageConversion(t *testing.T) {
	tests := []struct {
		name        string
		messages    []Message
		tools       json.RawMessage
		effort      string
		validateReq func(t *testing.T, body map[string]any)
	}{
		{
			name: "system and image conversion",
			messages: []Message{
				{Role: "system", Content: "sys prompt"},
				{Role: "user", Content: "see", Images: []Image{{MimeType: "image/png", Data: "QUJD"}}},
			},
			validateReq: func(t *testing.T, body map[string]any) {
				if body["system"] != "sys prompt" || int(body["max_tokens"].(float64)) != 32000 {
					t.Fatalf("bad system/max_tokens: %+v", body)
				}
				msgs := body["messages"].([]any)
				if len(msgs) != 1 {
					t.Fatalf("expected 1 msg, got %d", len(msgs))
				}
				user := msgs[0].(map[string]any)
				blocks := user["content"].([]any)
				if len(blocks) != 2 || blocks[0].(map[string]any)["type"] != "text" || blocks[1].(map[string]any)["type"] != "image" {
					t.Fatalf("bad content blocks: %+v", blocks)
				}
			},
		},
		{
			name: "assistant tool call and merged consecutive tool results",
			messages: []Message{
				{
					Role: "assistant",
					ToolCalls: []ToolCall{{ID: "c1", Type: "function", Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{Name: "read", Arguments: `{"path":"x"}`}}},
				},
				{Role: "tool", ToolCallID: "c1", Content: "res1"},
				{Role: "tool", ToolCallID: "c2", Content: "res2"},
			},
			tools:  json.RawMessage(`[{"type":"function","function":{"name":"read","description":"d","parameters":{"type":"object"}}}]`),
			effort: "high",
			validateReq: func(t *testing.T, body map[string]any) {
				if th, ok := body["thinking"].(map[string]any); !ok || th["type"] != "adaptive" {
					t.Fatalf("missing thinking: %+v", body["thinking"])
				}
				if oc, ok := body["output_config"].(map[string]any); !ok || oc["effort"] != "high" {
					t.Fatalf("missing effort: %+v", body["output_config"])
				}
				tools := body["tools"].([]any)
				if len(tools) != 1 || tools[0].(map[string]any)["name"] != "read" {
					t.Fatalf("bad tools: %+v", tools)
				}
				msgs := body["messages"].([]any)
				if len(msgs) != 2 {
					t.Fatalf("expected 2 msgs, got %d", len(msgs))
				}
				toolRes := msgs[1].(map[string]any)["content"].([]any)
				if len(toolRes) != 2 || toolRes[0].(map[string]any)["type"] != "tool_result" {
					t.Fatalf("bad merged tool results: %+v", toolRes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var received map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != "secret" || r.Header.Get("anthropic-version") != "2023-06-01" {
					t.Errorf("unexpected headers/path: %s %v", r.URL.Path, r.Header)
				}
				data, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(data, &received)
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
			}))
			defer server.Close()
			client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "secret"})
			_, _ = client.Stream(t.Context(), "claude", tt.effort, tt.messages, tt.tools, func(StreamEvent) {})
			tt.validateReq(t, received)
		})
	}
}
