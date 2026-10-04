package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaxTokensLimit(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{`{"message":"max_tokens: 32000 > 8192, which is the maximum allowed number of output tokens for claude-3-haiku-20240307"}`, 8192},
		{`max_tokens must be less than or equal to 4096`, 4096},
		{`"max_tokens" must be <= 16384`, 16384},
		{`input length and max_tokens exceed context limit: 188240 + 32000 > 200000`, 0},
		{`max_tokens: 32000 > 64000`, 0},
		{`max_tokens: Field required`, 0},
		{`thinking not supported`, 0},
	}
	for _, tt := range tests {
		got, ok := maxTokensLimit(tt.text, 32000)
		if got != tt.want || ok != (tt.want > 0) {
			t.Errorf("%q: got %d %v, want %d", tt.text, got, ok, tt.want)
		}
	}
}

func TestAnthropicMaxTokensRetry(t *testing.T) {
	var seen []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var body struct {
			MaxTokens int             `json:"max_tokens"`
			Thinking  json.RawMessage `json:"thinking"`
		}
		_ = json.Unmarshal(data, &body)
		seen = append(seen, body.MaxTokens)
		if body.MaxTokens > 8192 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"max_tokens: 32000 > 8192"}}`))
			return
		}
		if body.Thinking == nil {
			t.Error("max_tokens retry must keep thinking")
		}
		_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{Kind: KindAnthropic, BaseURL: server.URL, APIKey: "test-key"})
	for range 2 {
		if _, err := client.Stream(t.Context(), "m", "high", nil, nil, func(StreamEvent) {}); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 3 || seen[0] != 32000 || seen[1] != 8192 || seen[2] != 8192 {
		t.Fatalf("max_tokens sequence: %v", seen)
	}
}
