package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const anthropicNativeSSE = `data: {"type":"message_start","message":{"role":"assistant","usage":{"input_tokens":0,"cache_read_input_tokens":80,"cache_creation_input_tokens":20}}}

data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"hmm"}}

data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"SIG"}}

data: {"type":"content_block_stop","index":0}

data: {"type":"content_block_start","index":1,"content_block":{"type":"redacted_thinking","data":"RED"}}

data: {"type":"content_block_stop","index":1}

data: {"type":"content_block_start","index":2,"content_block":{"type":"tool_use","id":"t1","name":"read","input":{}}}

data: {"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"{\"path\":\"a\"}"}}

data: {"type":"content_block_stop","index":2}

data: {"type":"message_delta","usage":{"output_tokens":9}}

data: {"type":"message_stop"}

`

func TestAnthropicNativeBlocksSurviveReplay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(anthropicNativeSSE))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "k"})
	got, err := client.Stream(t.Context(), "c", "high", []Message{{Role: "user", Content: "hi"}}, nil, func(StreamEvent) {})
	if err != nil {
		t.Fatal(err)
	}
	u := got.Usage
	if u.Input != 100 || u.CachedInput != 80 || u.CacheWriteInput != 20 || !u.CacheKnown || !u.CacheWriteKnown || u.Output != 9 {
		t.Fatalf("usage: %+v", u)
	}
	n := got.Message.Native
	if n == nil || n.Model != "c" || len(n.Items) != 3 {
		t.Fatalf("native: %+v", n)
	}
	stored, _ := json.Marshal(got.Message)
	var reloaded Message
	if err := json.Unmarshal(stored, &reloaded); err != nil {
		t.Fatal(err)
	}
	history := []Message{{Role: "user", Content: "hi"}, reloaded,
		{Role: "tool", ToolCallID: "t1", Content: "body"}}
	data, err := buildAnthropicPayload("c", "high", history, nil, anthropicOptions{thinking: true, cache: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	order := []string{`"signature":"SIG"`, `"data":"RED"`, `"input":{"path":"a"}`, `"tool_use_id":"t1"`}
	last := -1
	for _, needle := range order {
		i := strings.Index(s, needle)
		if i <= last {
			t.Fatalf("missing or out of order %s in %s", needle, s)
		}
		last = i
	}
}

func TestAnthropicMissingCacheFieldsUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`data: {"type":"message_start","message":{"usage":{"input_tokens":7}}}` + "\n\n" +
			`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "k"})
	got, err := client.Stream(t.Context(), "c", "", nil, nil, func(StreamEvent) {})
	if err != nil || got.Usage.Input != 7 || got.Usage.CacheKnown || got.Usage.CacheWriteKnown {
		t.Fatalf("usage: %+v %v", got.Usage, err)
	}
}
