package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAutoUsesResponsesThenRemembersChat(t *testing.T) {
	var responses, chat atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/responses":
			responses.Add(1)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":"unsupported_endpoint","message":"The /responses endpoint is not supported"}}`))
		case "/chat/completions":
			chat.Add(1)
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n"))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	c := NewClient(Config{Kind: KindAuto, BaseURL: server.URL, APIKey: "k"})
	if _, err := c.Stream(context.Background(), "m", "", nil, json.RawMessage(`[]`), func(StreamEvent) {}); err != nil {
		t.Fatal(err)
	}
	if got := c.ActualProtocol("m"); got != KindOpenAI || responses.Load() != 1 || chat.Load() != 1 {
		t.Fatalf("protocol=%q requests=%d/%d", got, responses.Load(), chat.Load())
	}
}

func TestProtocolClassifierFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   bool
	}{
		{404, `{"error":{"code":"unsupported_endpoint","message":"The /responses endpoint is not supported"}}`, true},
		{400, `{"error":{"message":"invalid api key"}}`, false},
		{400, `{"error":{"message":"unsupported parameter reasoning.effort"}}`, false},
		{500, `{"error":{"code":"unsupported_endpoint"}}`, false},
	} {
		if got := protocolRejected(&statusError{status: tc.status, body: []byte(tc.body), protocol: KindResponses}); got != tc.want {
			t.Errorf("status=%d body=%s got=%v want=%v", tc.status, tc.body, got, tc.want)
		}
	}
}
