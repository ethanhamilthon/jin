package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAnthropicErrorKeepsReportedUsage(t *testing.T) {
	sse := "data: {\"type\":\"message_start\",\"message\":{\"role\":\"assistant\",\"usage\":{\"input_tokens\":12}}}\n\n" +
		"data: {\"type\":\"error\",\"error\":{\"type\":\"overloaded_error\",\"message\":\"busy\"}}\n\n"
	response, err := parseAnthropicStream(strings.NewReader(sse), func(StreamEvent) {}, func([]byte) {})
	if err == nil || !response.Usage.Known || response.Usage.Input != 12 {
		t.Fatalf("response %+v err %v", response, err)
	}
}

func TestRetryAfterReadFailureReportsUsage(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) == 1 {
			chunk := "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"half\"}}],\"usage\":{\"prompt_tokens\":9,\"completion_tokens\":1}}\n\n"
			w.Header().Set("Content-Length", "100000")
			_, _ = w.Write([]byte(chunk))
			return
		}
		_, _ = w.Write([]byte(okStream))
	}))
	defer server.Close()
	var reset StreamEvent
	client := NewClient(Config{BaseURL: server.URL, APIKey: "k"})
	_, err := client.Stream(t.Context(), "m", "", nil, json.RawMessage("[]"), func(e StreamEvent) {
		if e.Kind == Reset {
			reset = e
		}
	})
	if err != nil || count != 2 || reset.Usage.Input != 9 || reset.Usage.Output != 1 {
		t.Fatalf("err %v count %d reset %+v", err, count, reset)
	}
}
