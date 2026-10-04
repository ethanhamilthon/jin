package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func anthropicFeatureServer(t *testing.T, reject bool, bodies *[]map[string]any) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(data, &body)
		*bodies = append(*bodies, body)
		if reject && body["thinking"] != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"thinking: unknown field"}}`))
			return
		}
		_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAnthropicFeaturesScopedByEndpoint(t *testing.T) {
	var proxyBodies, directBodies []map[string]any
	proxy := anthropicFeatureServer(t, true, &proxyBodies)
	direct := anthropicFeatureServer(t, false, &directBodies)
	client := NewClient(Config{Kind: KindAnthropic, BaseURL: proxy.URL, APIKey: "test-key"})
	if _, err := client.Stream(t.Context(), "m", "high", nil, nil, func(StreamEvent) {}); err != nil {
		t.Fatal(err)
	}
	if len(proxyBodies) != 2 {
		t.Fatalf("proxy requests=%d", len(proxyBodies))
	}
	client.Configure(Config{Kind: KindAnthropic, BaseURL: direct.URL, APIKey: "test-key"})
	if _, err := client.Stream(t.Context(), "m", "high", nil, nil, func(StreamEvent) {}); err != nil {
		t.Fatal(err)
	}
	if len(directBodies) != 1 || directBodies[0]["thinking"] == nil {
		t.Fatalf("direct endpoint must keep thinking: %+v", directBodies)
	}
	client.Configure(Config{Kind: KindAnthropic, BaseURL: proxy.URL, APIKey: "test-key"})
	if _, err := client.Stream(t.Context(), "m", "high", nil, nil, func(StreamEvent) {}); err != nil {
		t.Fatal(err)
	}
	if len(proxyBodies) != 3 || proxyBodies[2]["thinking"] != nil {
		t.Fatalf("proxy must remember disabled thinking: %d", len(proxyBodies))
	}
}
