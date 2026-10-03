package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func chatUsage(t *testing.T, usage string) Usage {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"role":"assistant","content":"x"}}]}` + "\n\n" +
			`data: {"choices":[],"usage":` + usage + "}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	got, err := NewClient(Config{BaseURL: server.URL, APIKey: "k"}).Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {})
	if err != nil {
		t.Fatal(err)
	}
	return got.Usage
}

func TestChatUsageBreakdown(t *testing.T) {
	u := chatUsage(t, `{"prompt_tokens":50,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":0},"completion_tokens_details":{"reasoning_tokens":2}}`)
	if u.Input != 50 || !u.CacheKnown || u.CachedInput != 0 || !u.ReasoningKnown || u.Reasoning != 2 {
		t.Fatalf("explicit zero: %+v", u)
	}
	u = chatUsage(t, `{"prompt_tokens":50,"completion_tokens":5}`)
	if u.CacheKnown || u.ReasoningKnown {
		t.Fatalf("missing details must stay unknown: %+v", u)
	}
}
