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

func keyOf(t *testing.T, data []byte) string {
	t.Helper()
	var body struct {
		Key string `json:"prompt_cache_key"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	return body.Key
}

func TestCacheKeyFollowsOnlyTheStablePart(t *testing.T) {
	a := []Message{{Role: "system", Content: "stable\n" + CacheBreak + "\nsession 1"}}
	b := []Message{{Role: "system", Content: "stable\n" + CacheBreak + "\nsession 2"}}
	c := []Message{{Role: "system", Content: "other project\n" + CacheBreak + "\nsession 1"}}
	if promptCacheKey(a) != promptCacheKey(b) || promptCacheKey(a) == promptCacheKey(c) {
		t.Fatalf("keys: %s %s %s", promptCacheKey(a), promptCacheKey(b), promptCacheKey(c))
	}
	if len(promptCacheKey(a)) != 16 || promptCacheKey(nil) != "" {
		t.Fatalf("key %q, empty %q", promptCacheKey(a), promptCacheKey(nil))
	}
}

func TestResponsesPayloadAlwaysHasTheKey(t *testing.T) {
	data, _ := buildResponsesPayload("m", "", splitMessages(), nil)
	if keyOf(t, data) != promptCacheKey(splitMessages()) || keyOf(t, data) == "" {
		t.Fatalf("payload: %s", data)
	}
	data, _ = buildResponsesPayload("m", "", []Message{{Role: "user", Content: "hi"}}, nil)
	if strings.Contains(string(data), "prompt_cache_key") {
		t.Fatalf("no system prompt, no key: %s", data)
	}
}

func TestChatRequestDropsRejectedCacheKeyAndRemembersIt(t *testing.T) {
	var seen []bool
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		has := keyOf(t, data) != ""
		seen = append(seen, has)
		calls.Add(1)
		if has {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"Unknown parameter: 'prompt_cache_key'"}}`))
			return
		}
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindOpenAI, BaseURL: server.URL, APIKey: "sk-test-secret"})
	for range 2 {
		resp, err := client.chatRequest(t.Context(), "m", "", splitMessages(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
	if len(seen) != 3 || !seen[0] || seen[1] || seen[2] {
		t.Fatalf("key sent per request: %v", seen)
	}
}

func TestChatRequestSendsTheKeyWhenAccepted(t *testing.T) {
	var key string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		key = keyOf(t, data)
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindOpenAI, BaseURL: server.URL, APIKey: "sk-test-secret"})
	resp, err := client.chatRequest(t.Context(), "m", "", splitMessages(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if key != promptCacheKey(splitMessages()) {
		t.Fatalf("key = %q", key)
	}
}
