package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func splitMessages() []Message {
	return []Message{
		{Role: "system", Content: "stable text\n" + CacheBreak + "\nvolatile text"},
		{Role: "user", Content: "hi"},
	}
}

func TestSplitSystem(t *testing.T) {
	cases := []struct{ in, stable, tail string }{
		{"all stable", "all stable", ""},
		{"a\n" + CacheBreak + "\nb", "a", "b"},
		{"a\n" + CacheBreak + "\n", "a", ""},
		{"a " + CacheBreak + " in repo\n" + CacheBreak + "\nb", "a  in repo", "b"},
		{"", "", ""},
	}
	for _, c := range cases {
		if stable, tail := SplitSystem(c.in); stable != c.stable || tail != c.tail {
			t.Errorf("SplitSystem(%q) = %q, %q; want %q, %q", c.in, stable, tail, c.stable, c.tail)
		}
	}
}

func TestAnthropicSystemIsTwoBlocksWithOneBreakpoint(t *testing.T) {
	data, err := buildAnthropicPayload("m", "", splitMessages(), nil, anthropicOptions{cache: true, maxTokens: 10})
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		System []struct {
			Text         string         `json:"text"`
			CacheControl map[string]any `json:"cache_control"`
		} `json:"system"`
		CacheControl map[string]any `json:"cache_control"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.System) != 2 || body.System[0].Text != "stable text" || body.System[1].Text != "volatile text" {
		t.Fatalf("system: %s", data)
	}
	if body.System[0].CacheControl["type"] != "ephemeral" || body.System[1].CacheControl != nil || body.CacheControl == nil {
		t.Fatalf("breakpoints: %s", data)
	}
	if strings.Contains(string(data), CacheBreak) {
		t.Fatalf("marker on the wire: %s", data)
	}
}

func TestAnthropicSystemWithoutMarkerStaysAString(t *testing.T) {
	data, _ := buildAnthropicPayload("m", "", []Message{{Role: "system", Content: "plain"}}, nil, anthropicOptions{cache: true, maxTokens: 10})
	if !strings.Contains(string(data), `"system":"plain"`) {
		t.Fatalf("payload: %s", data)
	}
}

func TestCacheControlFallbackDropsTheBlockField(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		if calls.Add(1) == 1 {
			if strings.Count(string(data), "cache_control") != 2 {
				t.Errorf("first request needs both breakpoints: %s", data)
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"bad cache_control"}}`))
			return
		}
		if strings.Contains(string(data), "cache_control") || !strings.Contains(string(data), "volatile text") {
			t.Errorf("retry must carry no cache_control: %s", data)
		}
		_, _ = w.Write([]byte(`data: {"type":"message_start","message":{"usage":{"input_tokens":1}}}` + "\n\n" +
			`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "k"})
	if _, err := client.Stream(t.Context(), "c", "", splitMessages(), nil, func(StreamEvent) {}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestOpenAIAdaptersJoinBothPartsWithoutTheMarker(t *testing.T) {
	chat, err := chatPayload("m", "", splitMessages(), nil)
	if err != nil {
		t.Fatal(err)
	}
	responses, err := buildResponsesPayload("m", "", splitMessages(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"chat": chat, "responses": responses} {
		if strings.Contains(string(data), CacheBreak) {
			t.Errorf("%s: marker on the wire: %s", name, data)
		}
		if !strings.Contains(string(data), `stable text\n\nvolatile text`) {
			t.Errorf("%s: parts not joined with a blank line: %s", name, data)
		}
	}
}
