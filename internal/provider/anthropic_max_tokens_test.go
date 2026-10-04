package provider

import (
	"encoding/json"
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
		{`max_tokens must not exceed 8192`, 8192},
		{`max_tokens must be >= 1 and <= 8192`, 8192},
		{`max_tokens must be less than or equal to 4096`, 4096},
		{`"max_tokens" must be <= 16384`, 16384},
		{`max_tokens cannot exceed 8192`, 8192},
		{`max_tokens must be between 1 and 8192`, 8192},
		{`max_tokens: at most 8192`, 8192},
		{`max_tokens: limit of 8192`, 8192},
		{`max_tokens must be less than 8192`, 8191},
		{`max_tokens must be < 8192`, 8191},
		{`max_tokens must be >= 1 and < 8192`, 8191},
		{`max_tokens must be below 4096`, 4095},
		{`input length and max_tokens exceed context limit: 188240 + 32000 > 200000`, 0},
		{`max_tokens: 32000 > 64000`, 0},
		{`max_tokens: Field required`, 0},
		{`max_tokens must be >= 1`, 0},
		{`max_tokens must be less than 1`, 0},
		{`thinking not supported`, 0},
	}
	for _, tt := range tests {
		got, ok := maxTokensLimit(tt.text, 32000)
		if got != tt.want || ok != (tt.want > 0) {
			t.Errorf("%q: got %d %v, want %d", tt.text, got, ok, tt.want)
		}
	}
}

func TestAnthropicMaxTokensRetryFormats(t *testing.T) {
	tests := []struct {
		name      string
		errMsg    string
		wantLimit int
	}{
		{"range format", `max_tokens must be >= 1 and <= 8192`, 8192},
		{"not exceed format", `max_tokens must not exceed 4096`, 4096},
		{"strict less than format", `max_tokens must be less than 8192`, 8191},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen []int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					MaxTokens int             `json:"max_tokens"`
					Thinking  json.RawMessage `json:"thinking"`
				}
				_ = json.NewDecoder(r.Body).Decode(&body)
				seen = append(seen, body.MaxTokens)
				if body.MaxTokens > tt.wantLimit {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":{"message":"` + tt.errMsg + `"}}`))
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
			if len(seen) != 3 || seen[0] != 32000 || seen[1] != tt.wantLimit || seen[2] != tt.wantLimit {
				t.Fatalf("sequence mismatch: got %v", seen)
			}
		})
	}
}
