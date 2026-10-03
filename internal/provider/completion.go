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
	Timeout: 30 * time.Minute,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type Usage struct {
	Input           int  `json:"input_tokens"`
	Output          int  `json:"output_tokens"`
	CachedInput     int  `json:"cached_input_tokens"`
	CacheWriteInput int  `json:"cache_write_input_tokens"`
	Reasoning       int  `json:"reasoning_tokens"`
	CacheKnown      bool `json:"cache_known"`
	CacheWriteKnown bool `json:"cache_write_known"`
	ReasoningKnown  bool `json:"reasoning_known"`
	Known           bool `json:"known"`
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
	}{model, chatMessages(messages), effort, toolsSchema, true, streamOptions{true}})
	if err != nil {
		return nil, errors.New("cannot encode chat request")
	}
	return payload, nil
}
