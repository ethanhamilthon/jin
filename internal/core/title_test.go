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

func TestTitleSendsOnlySubjectAndFixedInstruction(t *testing.T) {
	client, fake := titleClient(t, "Parser bug")
	got, err := Title(context.Background(), client, "m", "", "Name it.")
	if err != nil || got != "Parser bug" {
		t.Fatalf("title = %q, %v", got, err)
	}
	sent := fake.sent(0)
	if len(sent) != 1 || sent[0].Role != "user" || sent[0].Content != titleInstruction+"<subject>\nName it.\n</subject>" {
		t.Fatalf("sent %+v", sent)
	}
}

func TestTitleNeedsModelAndText(t *testing.T) {
	client, _ := titleClient(t, "  ")
	if _, err := Title(context.Background(), client, "", "", "Name it."); err == nil {
		t.Error("no model was accepted")
	}
	if _, err := Title(context.Background(), client, "m", "", "Name it."); err == nil {
		t.Error("an empty answer was accepted")
	}
}
