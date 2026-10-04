package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

// fakeProvider answers each request with the next scripted reply: "overflow"
// is an HTTP 400 context_length_exceeded, "tool" is a call of the read tool,
// anything else is a text answer.
// The last reply repeats once the script runs out.
type fakeProvider struct {
	mu       sync.Mutex
	replies  []string
	requests [][]provider.Message
}

func newFakeAgent(t *testing.T, replies ...string) (*Agent, *fakeProvider) {
	fake := &fakeProvider{replies: replies}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	return NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry()), fake
}

func (f *fakeProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Messages []provider.Message `json:"messages"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	reply := f.replies[min(len(f.requests), len(f.replies)-1)]
	f.requests = append(f.requests, body.Messages)
	f.mu.Unlock()
	if reply == "overflow" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"context_length_exceeded","message":"too long"}}`))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	if reply == "tool" {
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0,\"id\":\"c1\",\"function\":{\"name\":\"read\",\"arguments\":\"{}\"}}]}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n")
		return
	}
	_, _ = fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":%q}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n", reply)
}

func (f *fakeProvider) sent(i int) []provider.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests[i]
}

func (f *fakeProvider) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

// toolTurns builds a history of n turns, each with one read call whose
// result has size bytes.
func toolTurns(n, size int) []provider.Message {
	history := []provider.Message{{Role: "system", Content: "sys"}}
	for i := range n {
		id := fmt.Sprintf("c%d", i)
		call := provider.ToolCall{ID: id, Type: "function"}
		call.Function.Name, call.Function.Arguments = "read", "{}"
		history = append(history,
			provider.Message{Role: "user", Content: fmt.Sprintf("prompt %d", i)},
			provider.Message{Role: "assistant", ToolCalls: []provider.ToolCall{call}},
			provider.Message{Role: "tool", ToolCallID: id, Content: strings.Repeat("x", size)},
			provider.Message{Role: "assistant", Content: "done"})
	}
	return history
}

func omitted(messages []provider.Message) int {
	n := 0
	for _, msg := range messages {
		if msg.Role == "tool" && strings.HasPrefix(msg.Content, omittedPrefix) {
			n++
		}
	}
	return n
}

func collect(updates chan Update) []Update {
	close(updates)
	var all []Update
	for u := range updates {
		all = append(all, u)
	}
	return all
}
