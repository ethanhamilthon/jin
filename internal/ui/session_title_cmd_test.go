package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"jin/internal/store"
)

// answerServer replies to every request with the title "Greeting chat.".
func answerServer(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"Greeting chat.\"}}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestTitleCommandNamesTheSessionInTheBackground(t *testing.T) {
	a, s := persistedApp(t)
	a.active = s
	a.sessions[s.id] = s
	if err := a.store.AddProvider(store.ProviderEntry{ID: "p", Name: "p", Kind: "openai", BaseURL: answerServer(t), APIKey: "k"}); err != nil {
		t.Fatal(err)
	}
	if err := a.store.SetActiveProvider("p"); err != nil {
		t.Fatal(err)
	}
	a.cfg, _ = a.store.LoadConfig()
	s.provider = "p"
	a.titles = make(chan titleResult, 1)
	a.nameNow()
	select {
	case r := <-a.titles:
		a.receiveTitle(r)
	case <-time.After(5 * time.Second):
		t.Fatal("the title was not generated")
	}
	if s.title != "Greeting chat" || lastText(s) != "Session named: Greeting chat" {
		t.Fatalf("title %q, last entry %q", s.title, lastText(s))
	}
	if rec, _, _ := a.store.GetSession(s.id); rec.Title != "Greeting chat" {
		t.Fatalf("stored title %q", rec.Title)
	}
}

func TestTitleWithoutAnEnabledProviderSendsNothing(t *testing.T) {
	a, s := persistedApp(t)
	a.active = s
	a.cfg.Providers = []store.ProviderEntry{{ID: "p", Name: "p", Disabled: true}}
	s.provider = "p"
	a.titles = make(chan titleResult, 1)
	a.nameNow()
	if lastText(s) != noEnabledText || len(a.titles) != 0 {
		t.Fatalf("last entry %q, results %d", lastText(s), len(a.titles))
	}
}

func TestTitleOfAnEmptySessionSaysSo(t *testing.T) {
	a, _ := layoutApp(t)
	a.nameNow()
	if lastText(a.active) != "Nothing to title yet: this session has no messages" {
		t.Fatalf("last entry %q", lastText(a.active))
	}
}

func TestSendWithoutAnEnabledProviderTriesNoRequest(t *testing.T) {
	a, s := persistedApp(t)
	a.active, s.ready = s, true
	a.cfg.Providers = []store.ProviderEntry{{ID: "p", Name: "p", Disabled: true}, {ID: "q", Name: "q", Disabled: true}}
	s.provider = "p"
	a.sendDraft("hello")
	if len(s.pending) != 0 || lastText(s) != noEnabledText {
		t.Fatalf("pending %d, last %q", len(s.pending), lastText(s))
	}
	if got := strings.Join(s.input, ""); got != "hello" {
		t.Errorf("the draft was not given back: %q", got)
	}
}

func TestSendInADisabledSessionProviderIsRefusedWhileAnotherIsOn(t *testing.T) {
	a, s := persistedApp(t)
	a.active, s.ready = s, true
	a.cfg.Providers = []store.ProviderEntry{{ID: "p", Name: "p", Disabled: true}, {ID: "q", Name: "q"}}
	s.provider = "p"
	a.sendDraft("hello")
	if len(s.pending) != 0 || lastText(s) != noEnabledText {
		t.Fatalf("pending %d, last %q", len(s.pending), lastText(s))
	}
}
