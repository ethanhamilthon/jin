package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// completionClient allows long-lived streaming responses (heavy reasoning
// effort can run for minutes); user-initiated cancellation goes through ctx.
var completionClient = &http.Client{
	Timeout: 10 * time.Minute,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type Usage struct {
	Input       int
	Output      int
	CachedInput int
	CacheKnown  bool
	Known       bool
}

type Response struct {
	Message Message
	Usage   Usage
}

func chatPayload(model, effort string, messages []Message, toolsSchema json.RawMessage) ([]byte, error) {
	type streamOptions struct {
		IncludeUsage bool `json:"include_usage"`
	}
	payload, err := json.Marshal(struct {
		Model         string          `json:"model"`
		Messages      []Message       `json:"messages"`
		Effort        string          `json:"reasoning_effort,omitempty"`
		Tools         json.RawMessage `json:"tools,omitempty"`
		Stream        bool            `json:"stream"`
		StreamOptions streamOptions   `json:"stream_options"`
	}{model, messages, effort, toolsSchema, true, streamOptions{true}})
	if err != nil {
		return nil, errors.New("cannot encode chat request")
	}
	return payload, nil
}
