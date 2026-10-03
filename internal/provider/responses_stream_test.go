package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func responsesServer(t *testing.T, bodies ...string) (*Client, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Errorf("path %s", r.URL.Path)
		}
		n := int(calls.Add(1)) - 1
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(bodies[min(n, len(bodies)-1)]))
	}))
	t.Cleanup(server.Close)
	return NewClient(Config{Kind: KindResponses, BaseURL: server.URL, APIKey: "k"}), &calls
}

const responsesDone = `data: {"type":"response.output_text.delta","delta":"he"}

data: {"type":"response.reasoning_summary_text.delta","delta":"plan"}

data: {"type":"response.completed","response":{"output":[` +
	`{"type":"reasoning","id":"rs","encrypted_content":"ENC-BLOB","summary":[{"type":"summary_text","text":"plan"}]},` +
	`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello"}]},` +
	`{"type":"function_call","call_id":"c1","name":"read","arguments":"{\"path\":\"a\"}"}],` +
	`"usage":{"input_tokens":100,"output_tokens":7,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":3}}}}

`

func TestResponsesStream(t *testing.T) {
	client, _ := responsesServer(t, responsesDone)
	var text string
	got, err := client.Stream(t.Context(), "m", "low", nil, nil, func(e StreamEvent) {
		if e.Kind == DeltaContent {
			text += e.Text
		}
	})
	if err != nil || text != "he" || got.Message.Content != "hello" || got.Message.ReasoningContent != "plan" {
		t.Fatalf("stream: %+v %q %v", got, text, err)
	}
	if len(got.Message.ToolCalls) != 1 || got.Message.ToolCalls[0].ID != "c1" {
		t.Fatalf("calls: %+v", got.Message.ToolCalls)
	}
	n := got.Message.Native
	if n == nil || n.Kind != KindResponses || n.Model != "m" || len(n.Items) != 3 || !strings.Contains(string(n.Items[0]), "ENC-BLOB") {
		t.Fatalf("native: %+v", n)
	}
	u := got.Usage
	if u.Input != 100 || u.Output != 7 || !u.CacheKnown || u.CachedInput != 0 || u.Reasoning != 3 || u.CacheWriteKnown {
		t.Fatalf("usage: %+v", u)
	}
}

func TestResponsesOutputItemFallbackAndMultiline(t *testing.T) {
	body := "data: {\"type\":\"response.output_item.done\",\n" +
		"data: \"item\":{\"type\":\"message\",\"content\":[{\"type\":\"refusal\",\"refusal\":\"no\"}]}}\n\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\n"
	client, _ := responsesServer(t, body)
	got, err := client.Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {})
	if err != nil || got.Message.Content != "no" || got.Usage.CacheKnown {
		t.Fatalf("fallback: %+v %v", got, err)
	}
}

func TestResponsesErrors(t *testing.T) {
	cases := map[string]string{
		"truncated":   "data: {\"type\":\"response.output_text.delta\",\"delta\":\"x\"}\n\n",
		"failed":      `data: {"type":"response.failed","response":{"error":{"code":"invalid_prompt","message":"bad"}}}` + "\n\n",
		"incomplete":  `data: {"type":"response.incomplete","response":{"incomplete_details":{"reason":"max_output_tokens"}}}` + "\n\n",
		"malformed":   "data: {nope\n\n",
		"unsupported": `data: {"type":"response.completed","response":{"output":[{"type":"web_search_call"}]}}` + "\n\n",
		"bad_args":    `data: {"type":"response.completed","response":{"output":[{"type":"function_call","call_id":"c","name":"r","arguments":"{"}]}}` + "\n\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			client, _ := responsesServer(t, body)
			if _, err := client.Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestResponsesRetriesServerError(t *testing.T) {
	failed := `data: {"type":"response.failed","response":{"error":{"code":"server_error","message":"x"}}}` + "\n\n"
	client, calls := responsesServer(t, failed, responsesDone)
	if _, err := client.Stream(t.Context(), "m", "", nil, json.RawMessage(testToolsSchema), func(StreamEvent) {}); err != nil || calls.Load() != 2 {
		t.Fatalf("retry: %v calls=%d", err, calls.Load())
	}
}
