package provider

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsContextOverflow(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"openai code", &statusError{status: 400, msg: `/chat/completions HTTP 400: {"error":{"code":"context_length_exceeded"}}`}, true},
		{"anthropic", &statusError{status: 400, msg: `/messages HTTP 400: {"error":{"message":"prompt is too long: 210000 tokens > 200000 maximum"}}`}, true},
		{"generic phrase", errors.New("This model's maximum context length is 128000 tokens"), true},
		{"context window", errors.New("input exceeds the context window of this model"), true},
		{"wrapped", fmt.Errorf("turn: %w", &statusError{status: 400, msg: "Prompt Is Too Long"}), true},
		{"413 with context", &statusError{status: 413, msg: "/v1 HTTP 413: request exceeds context"}, true},
		{"413 without context", &statusError{status: 413, msg: "/v1 HTTP 413: payload too large"}, false},
		{"other 400", &statusError{status: 400, msg: "/v1 HTTP 400: invalid model"}, false},
		{"rate limit", &statusError{status: 429, msg: "/v1 HTTP 429: slow down"}, false},
	}
	for _, c := range cases {
		if got := IsContextOverflow(c.err); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
