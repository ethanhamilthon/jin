package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicMaxTokensOnceOnly(t *testing.T) {
	var count int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusBadRequest)
		msg := "max_tokens: 32000 > 8192"
		if count > 1 {
			msg = "max_tokens: 8192 > 4096"
		}
		_, _ = w.Write([]byte(`{"error":{"message":"` + msg + `"}}`))
	}))
	defer server.Close()

	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "test-key"})
	if _, err := client.Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {}); err == nil {
		t.Fatal("expected error after second 400 rejection")
	}
	if count != 2 {
		t.Fatalf("expected exactly 2 requests (1 retry only), got %d", count)
	}
}

func TestAnthropicMaxTokensEndpointSwitch(t *testing.T) {
	var proxySeen, directSeen []int
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			MaxTokens int `json:"max_tokens"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		proxySeen = append(proxySeen, b.MaxTokens)
		if b.MaxTokens > 8192 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"max_tokens must not exceed 8192"}}`))
			return
		}
		_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer proxy.Close()
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			MaxTokens int `json:"max_tokens"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		directSeen = append(directSeen, b.MaxTokens)
		_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer direct.Close()

	client := NewClient(Config{Kind: KindAnthropic, BaseURL: proxy.URL, APIKey: "test-key"})
	_, _ = client.Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {})
	client.Configure(Config{Kind: KindAnthropic, BaseURL: direct.URL, APIKey: "test-key"})
	_, _ = client.Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {})
	client.Configure(Config{Kind: KindAnthropic, BaseURL: proxy.URL, APIKey: "test-key"})
	_, _ = client.Stream(t.Context(), "m", "", nil, nil, func(StreamEvent) {})

	if len(proxySeen) != 3 || proxySeen[0] != 32000 || proxySeen[1] != 8192 || proxySeen[2] != 8192 {
		t.Fatalf("unexpected proxy requests: %v", proxySeen)
	}
	if len(directSeen) != 1 || directSeen[0] != 32000 {
		t.Fatalf("direct endpoint should receive default 32000, got: %v", directSeen)
	}
}

func TestAnthropicMaxTokensModelIsolation(t *testing.T) {
	var seen = make(map[string][]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Model     string `json:"model"`
			MaxTokens int    `json:"max_tokens"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		seen[b.Model] = append(seen[b.Model], b.MaxTokens)
		if b.Model == "haiku" && b.MaxTokens > 8192 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"max_tokens: 32000 > 8192"}}`))
			return
		}
		_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer server.Close()

	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "test-key"})
	_, _ = client.Stream(t.Context(), "haiku", "", nil, nil, func(StreamEvent) {})
	_, _ = client.Stream(t.Context(), "opus", "", nil, nil, func(StreamEvent) {})

	if len(seen["haiku"]) != 2 || seen["haiku"][0] != 32000 || seen["haiku"][1] != 8192 {
		t.Fatalf("haiku requests unexpected: %v", seen["haiku"])
	}
	if len(seen["opus"]) != 1 || seen["opus"][0] != 32000 {
		t.Fatalf("opus must keep default 32000, got: %v", seen["opus"])
	}
}
