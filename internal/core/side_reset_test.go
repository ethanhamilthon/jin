package core

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func TestSideRequestCountsUsageOfRetriedAttempt(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if atomic.AddInt32(&count, 1) == 1 {
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"half\"}}],\"usage\":{\"prompt_tokens\":7,\"completion_tokens\":2}}\n\n"))
			return
		}
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"summary\"}}],\"usage\":{\"prompt_tokens\":9,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "system", Content: "sys"}, {Role: "user", Content: "hi"}}
	updates := make(chan Update, 16)
	text, _, err := agent.sideRequest(t.Context(), t.Context(), Request{Model: "m"}, history, "instruction", updates)
	if err != nil || text != "summary" {
		t.Fatalf("text %q err %v", text, err)
	}
	var failed Update
	for _, u := range collect(updates) {
		if u.Kind == UpdateUsage && failed.Kind == "" {
			failed = u
		}
	}
	if failed.Usage.Input != 7 || failed.Usage.Output != 2 || failed.Model != "m" {
		t.Errorf("failed attempt usage = %+v", failed)
	}
}
