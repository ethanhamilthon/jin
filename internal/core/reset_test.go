package core

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func TestRetriedStreamSendsResetWithFailedUsage(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if atomic.AddInt32(&count, 1) == 1 {
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"half\"}}],\"usage\":{\"prompt_tokens\":7,\"completion_tokens\":2}}\n\n"))
			return
		}
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"whole\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "user", Content: "hi"}}
	updates := make(chan Update, 64)
	if err := agent.answer(t.Context(), t.Context(), Request{Model: "m"}, &history, nil, updates); err != nil {
		t.Fatal(err)
	}
	var kinds []UpdateKind
	var reset Update
	for _, u := range collect(updates) {
		if u.Kind == UpdateReset {
			reset = u
		}
		if u.Kind != UpdateHistory {
			kinds = append(kinds, u.Kind)
		}
	}
	want := []UpdateKind{UpdateAssistantDelta, UpdateReset, UpdateInfo, UpdateAssistantDelta, UpdateUsage}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("kinds = %v, want %v", kinds, want)
		}
	}
	if reset.Usage.Input != 7 || reset.Usage.Output != 2 || !reset.Usage.Known || reset.Model != "m" {
		t.Errorf("reset = %+v", reset)
	}
}
