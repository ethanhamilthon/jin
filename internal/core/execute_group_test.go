package core

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"jin/internal/provider"
	"jin/internal/tools"
)

type span struct{ start, end time.Time }

type slowTool struct {
	name  string
	mu    *sync.Mutex
	spans map[string]span
}

func (s slowTool) Name() string                  { return s.name }
func (slowTool) Schema() json.RawMessage         { return json.RawMessage(`{}`) }
func (slowTool) Summary(a string) (string, bool) { return a, true }
func (s slowTool) Run(ctx context.Context, a string) (string, error) {
	start := time.Now()
	select {
	case <-time.After(60 * time.Millisecond):
	case <-ctx.Done():
	}
	s.mu.Lock()
	s.spans[a] = span{start, time.Now()}
	s.mu.Unlock()
	return s.name + a, nil
}

func fakeCall(id, name, args string) provider.ToolCall {
	call := provider.ToolCall{ID: id, Type: "function"}
	call.Function.Name, call.Function.Arguments = name, args
	return call
}

func runFake(t *testing.T, calls []provider.ToolCall) (map[string]span, []provider.Message) {
	t.Helper()
	mu, spans := &sync.Mutex{}, map[string]span{}
	registry := tools.NewRegistry(slowTool{"read", mu, spans}, slowTool{"write", mu, spans})
	agent := NewAgent(nil, "sys", registry)
	var history []provider.Message
	if err := agent.runTools(t.Context(), t.Context(), Request{}, calls, &history, make(chan Update, 64)); err != nil {
		t.Fatal(err)
	}
	return spans, history
}

func overlaps(a, b span) bool { return a.start.Before(b.end) && b.start.Before(a.end) }

func TestReadsRunConcurrentlyAndKeepOrder(t *testing.T) {
	spans, history := runFake(t, []provider.ToolCall{fakeCall("1", "read", `"a"`), fakeCall("2", "read", `"b"`), fakeCall("3", "read", `"c"`)})
	if !overlaps(spans[`"a"`], spans[`"b"`]) || !overlaps(spans[`"b"`], spans[`"c"`]) {
		t.Fatalf("reads did not overlap: %+v", spans)
	}
	for i, id := range []string{"1", "2", "3"} {
		if history[i].ToolCallID != id {
			t.Fatalf("order broken: %+v", history)
		}
	}
}

func TestCallsOfOneAnswerRunTogether(t *testing.T) {
	spans, history := runFake(t, []provider.ToolCall{fakeCall("1", "read", `"a"`), fakeCall("2", "write", `"w"`), fakeCall("3", "read", `"b"`)})
	if !overlaps(spans[`"a"`], spans[`"w"`]) || !overlaps(spans[`"w"`], spans[`"b"`]) {
		t.Fatalf("calls did not overlap: %+v", spans)
	}
	if len(history) != 3 || history[1].Content != `write"w"` {
		t.Fatalf("history = %+v", history)
	}
}

func TestCancelStopsAllRunningReads(t *testing.T) {
	mu, spans := &sync.Mutex{}, map[string]span{}
	agent := NewAgent(nil, "sys", tools.NewRegistry(slowTool{"read", mu, spans}))
	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(10*time.Millisecond, cancel)
	var history []provider.Message
	start := time.Now()
	_ = agent.runTools(ctx, t.Context(), Request{}, []provider.ToolCall{fakeCall("1", "read", `"a"`), fakeCall("2", "read", `"b"`)}, &history, make(chan Update, 16))
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("cancellation did not stop the reads")
	}
}
