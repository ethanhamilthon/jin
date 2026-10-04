package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const chatDone = `data: {"choices":[{"delta":{"role":"assistant","content":"ok"}}]}` + "\n\n" + "data: [DONE]\n\n"

func chatServer(t *testing.T, reject bool, seen *[]bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		has := keyOf(t, data) != ""
		*seen = append(*seen, has)
		if has && reject {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"Unknown parameter: 'prompt_cache_key'"}}`))
			return
		}
		_, _ = w.Write([]byte(chatDone))
	}))
}

func TestClientStreamSendsCacheKeyOnChatCompletions(t *testing.T) {
	var seen []bool
	server := chatServer(t, false, &seen)
	defer server.Close()
	client := NewClient(Config{Kind: KindOpenAI, BaseURL: server.URL, APIKey: "sk-test-secret"})
	resp, err := client.Stream(t.Context(), "m", "", splitMessages(), nil, func(StreamEvent) {})
	if err != nil || resp.Message.Content != "ok" {
		t.Fatalf("resp %+v, err %v", resp, err)
	}
	if len(seen) != 1 || !seen[0] {
		t.Fatalf("key sent: %v", seen)
	}
}

func TestClientStreamDropsRejectedCacheKeyAndRemembersIt(t *testing.T) {
	var seen []bool
	server := chatServer(t, true, &seen)
	defer server.Close()
	client := NewClient(Config{Kind: KindOpenAI, BaseURL: server.URL, APIKey: "sk-test-secret"})
	for range 2 {
		if _, err := client.Stream(t.Context(), "m", "", splitMessages(), nil, func(StreamEvent) {}); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 3 || !seen[0] || seen[1] || seen[2] {
		t.Fatalf("key sent per request: %v", seen)
	}
}
