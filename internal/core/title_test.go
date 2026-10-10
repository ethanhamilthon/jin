package core

import (
	"context"
	"net/http/httptest"
	"testing"

	"jin/internal/provider"
)

func titleClient(t *testing.T, replies ...string) (*provider.Client, *fakeProvider) {
	fake := &fakeProvider{replies: replies}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	return provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), fake
}

func TestTitleSendsHistoryAndPrompt(t *testing.T) {
	client, fake := titleClient(t, "Parser bug")
	history := []provider.Message{{Role: "user", Content: "why does the parser fail"}, {Role: "assistant", Content: "a bug"}}
	got, err := Title(context.Background(), client, "m", "", history, nil, "Name it.")
	if err != nil || got != "Parser bug" {
		t.Fatalf("title = %q, %v", got, err)
	}
	sent := fake.sent(0)
	if len(sent) != 3 || sent[0].Content != history[0].Content || sent[2].Role != "user" || sent[2].Content != "Name it." {
		t.Fatalf("sent %+v", sent)
	}
	if len(history) != 2 {
		t.Fatal("history was changed")
	}
}

func TestTitleNeedsModelAndText(t *testing.T) {
	client, _ := titleClient(t, "  ")
	if _, err := Title(context.Background(), client, "", "", nil, nil, "Name it."); err == nil {
		t.Error("no model was accepted")
	}
	if _, err := Title(context.Background(), client, "m", "", nil, nil, "Name it."); err == nil {
		t.Error("an empty answer was accepted")
	}
}
