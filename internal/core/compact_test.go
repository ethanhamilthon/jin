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

func TestCompactRoundTrip(t *testing.T) {
	var sent []provider.Message
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []provider.Message `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		sent = body.Messages
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"the summary\"}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "system", Content: "sys"}, {Role: "user", Content: "hi"}, {Role: "assistant", Content: "hello"}}
	updates := make(chan Update, 8)
	if err := agent.compact(t.Context(), t.Context(), Request{Kind: RequestCompact, Model: "m"}, &history, updates); err != nil {
		t.Fatal(err)
	}
	if last := sent[len(sent)-1]; last.Role != "user" || !strings.HasPrefix(last.Content, "The conversation has grown long") || len(sent) != 4 {
		t.Fatalf("side request = %+v", sent)
	}
	if len(history) != 2 || history[0].Role != "system" || !IsSummary(history[1]) || !strings.Contains(history[1].Content, "the summary") {
		t.Fatalf("history = %+v", history)
	}
	if agent.size != 3 {
		t.Errorf("size = %d, want the summary output size 3", agent.size)
	}
	close(updates)
	var compacted bool
	for u := range updates {
		compacted = compacted || u.Kind == UpdateCompacted
	}
	if !compacted {
		t.Error("no UpdateCompacted sent")
	}
}

func TestDoneIsFinalOnlyForAnsweredPrompts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry())
	cases := []struct {
		request Request
		final   bool
	}{
		{Request{Prompt: "hi", Model: "m"}, true},
		{Request{Prompt: "hi"}, false},
		{Request{Kind: RequestCompact, Model: "m"}, false},
	}
	for _, c := range cases {
		history := []provider.Message{{Role: "system", Content: "sys"}, {Role: "user", Content: "earlier"}}
		updates := make(chan Update, 32)
		agent.turn(t.Context(), c.request, &history, nil, updates)
		close(updates)
		var done Update
		for u := range updates {
			if u.Kind == UpdateDone {
				done = u
			}
		}
		if done.Kind != UpdateDone || done.Final != c.final {
			t.Errorf("%+v: done = %+v, want Final=%v", c.request, done, c.final)
		}
	}
}
