package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func TestSideRequestRetriesOnToolCall(t *testing.T) {
	var attempts int
	var prompts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []provider.Message `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		prompts = append(prompts, body.Messages[len(body.Messages)-1].Content)
		w.Header().Set("Content-Type", "text/event-stream")
		attempts++
		if attempts == 1 {
			chunk := "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0,\"id\":\"c1\",\"function\":{\"name\":\"read\",\"arguments\":\"{}\"}}]}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n"
			_, _ = w.Write([]byte(chunk))
			return
		}
		chunk := "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"final summary\"}}],\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":5}}\n\ndata: [DONE]\n\n"
		_, _ = w.Write([]byte(chunk))
	}))
	defer server.Close()

	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "system", Content: "sys"}, {Role: "user", Content: "hi"}}
	updates := make(chan Update, 8)

	text, usage, err := agent.sideRequest(t.Context(), t.Context(), Request{Model: "m"}, history, "instruction", updates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "final summary" {
		t.Errorf("text = %q, want %q", text, "final summary")
	}
	if usage.Output != 5 {
		t.Errorf("usage.Output = %d, want 5", usage.Output)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
	if len(prompts) != 2 || prompts[0] != "instruction" || !strings.Contains(prompts[1], sideRetryReminder) {
		t.Errorf("prompts = %v", prompts)
	}
	close(updates)
	var updateUsages int
	for u := range updates {
		if u.Kind == UpdateUsage {
			updateUsages++
		}
	}
	if updateUsages != 1 {
		t.Errorf("got %d UpdateUsage from retry, want 1", updateUsages)
	}
}

func TestSideRequestFailsAfterSecondToolCall(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		attempts++
		chunk := "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0,\"id\":\"c1\",\"function\":{\"name\":\"read\",\"arguments\":\"{}\"}}]}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n"
		_, _ = w.Write([]byte(chunk))
	}))
	defer server.Close()

	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "system", Content: "sys"}}
	updates := make(chan Update, 8)

	_, _, err := agent.sideRequest(t.Context(), t.Context(), Request{Model: "m"}, history, "instruction", updates)
	if err == nil || !strings.Contains(err.Error(), "no text") {
		t.Fatalf("want 'no text' error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}
