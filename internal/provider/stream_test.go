package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamLocalServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("unexpected request: %s %s", r.URL.Path, r.Header.Get("Authorization"))
		}
		var body struct {
			Messages []Message `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Messages) != 1 {
			t.Errorf("unexpected messages: %+v %v", body, err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"hello\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{BaseURL: server.URL, APIKey: "secret"})
	var fragments []string
	response, err := client.Stream(t.Context(), "test", "", []Message{{Role: "user", Content: "hi"}}, json.RawMessage("[]"), func(e StreamEvent) { fragments = append(fragments, e.Text) })
	if err != nil || response.Message.Content != "hello" || strings.Join(fragments, "") != "hello" {
		t.Fatalf("stream: %+v %v %v", response, fragments, err)
	}
}

func TestStreamRejectsTruncatedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"partial\"}}]}\n\n"))
	}))
	defer server.Close()
	client := NewClient(Config{BaseURL: server.URL, APIKey: "secret"})
	if _, err := client.Stream(t.Context(), "test", "", nil, json.RawMessage("[]"), func(StreamEvent) {}); err == nil {
		t.Fatal("truncated stream must fail")
	}
}
