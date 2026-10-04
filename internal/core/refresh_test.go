package core

import (
	"context"
	"strings"
	"testing"
	"time"

	"jin/internal/provider"
)

var noon = time.Date(2026, 10, 4, 12, 0, 0, 0, time.Local)

type fakeClock struct{ now time.Time }

func (c *fakeClock) read() time.Time { return c.now }

func refreshAgent(t *testing.T, texts ...string) (*Agent, *fakeClock, *int) {
	t.Helper()
	agent, _ := newFakeAgent(t, "ok")
	clock := &fakeClock{now: noon}
	agent.clock = clock.read
	agent.SetSystemPrompt("sys")
	calls := 0
	agent.SetRefresher(func(context.Context) string {
		calls++
		return texts[min(calls-1, len(texts)-1)]
	})
	return agent, clock, &calls
}

func TestSystemStaysWhileTheCacheIsHotOrThePromptIsYoung(t *testing.T) {
	cases := []struct {
		name         string
		idle, age    time.Duration
		lastDoneSeen bool
		want         bool
	}{
		{"hot cache, old prompt", 4 * time.Minute, 3 * time.Hour, true, false},
		{"cold cache, young prompt", 10 * time.Minute, 30 * time.Minute, true, false},
		{"cold cache, old prompt", 10 * time.Minute, 2 * time.Hour, true, true},
		{"exactly at the TTL", 5 * time.Minute, 2 * time.Hour, true, false},
		{"no request yet, old prompt", 0, 2 * time.Hour, false, true},
	}
	for _, c := range cases {
		agent, clock, _ := refreshAgent(t, "new")
		clock.now = noon.Add(c.age)
		agent.renderedAt = noon
		if c.lastDoneSeen {
			agent.lastDone = clock.now.Add(-c.idle)
		}
		if got := agent.systemStale(clock.now); got != c.want {
			t.Errorf("%s: stale = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestSystemRefreshesWhenTheDateChanged(t *testing.T) {
	agent, clock, _ := refreshAgent(t, "new")
	agent.renderedAt = time.Date(2026, 10, 4, 23, 50, 0, 0, time.Local)
	agent.lastDone = agent.renderedAt
	clock.now = time.Date(2026, 10, 5, 0, 10, 0, 0, time.Local)
	if !agent.systemStale(clock.now) {
		t.Error("a new date after a cold cache must refresh")
	}
	clock.now = time.Date(2026, 10, 5, 0, 12, 0, 0, time.Local)
	agent.lastDone = clock.now.Add(-time.Minute)
	if agent.systemStale(clock.now) {
		t.Error("a hot cache must never refresh")
	}
}

func TestRefreshReplacesTheSystemAndNotesOnlyAChange(t *testing.T) {
	agent, clock, calls := refreshAgent(t, "sys", "changed")
	history := []provider.Message{{Role: "system", Content: "sys"}}
	agent.lastDone = noon
	clock.now = noon.Add(2 * time.Hour)
	if note := agent.refreshSystem(context.Background(), history); note != "" || *calls != 1 {
		t.Fatalf("same text: note %q, calls %d", note, *calls)
	}
	if agent.renderedAt != clock.now {
		t.Error("the render time must move even when the text is the same")
	}
	clock.now = clock.now.Add(2 * time.Hour)
	note := agent.refreshSystem(context.Background(), history)
	if !strings.Contains(note, "<system-refreshed>instructions were refreshed</system-refreshed>") {
		t.Errorf("note = %q", note)
	}
	if history[0].Content != "changed" || agent.systemPrompt != "changed" {
		t.Errorf("system = %q", history[0].Content)
	}
	if again := agent.refreshSystem(context.Background(), history); again != "" || *calls != 2 {
		t.Errorf("no second refresh right after: note %q, calls %d", again, *calls)
	}
}

func TestCancelledRefreshKeepsTheOldPrompt(t *testing.T) {
	agent, clock, _ := refreshAgent(t, "half [command cancelled]")
	history := []provider.Message{{Role: "system", Content: "sys"}}
	agent.lastDone = noon
	clock.now = noon.Add(2 * time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if note := agent.refreshSystem(ctx, history); note != "" || history[0].Content != "sys" {
		t.Errorf("note %q, system %q", note, history[0].Content)
	}
}

func TestRunRefreshesAfterCompactionAndOnAColdRequest(t *testing.T) {
	agent, fake := newFakeAgent(t, "answer one", "summary", "answer two", "answer three")
	clock := &fakeClock{now: noon}
	agent.clock = clock.read
	agent.SetSystemPrompt("sys")
	agent.SetRefresher(func(context.Context) string { return "sys v2" })
	prompts, updates := make(chan Request), make(chan Update, 64)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go agent.Run(ctx, nil, prompts, updates)
	step := func(request Request) {
		prompts <- request
		for update := range updates {
			if update.Kind == UpdateDone {
				return
			}
		}
	}
	step(Request{Prompt: "one", Model: "m"})
	step(Request{Kind: RequestCompact, Model: "m"})
	step(Request{Prompt: "two", Model: "m"})
	last := fake.sent(fake.count() - 1)
	if last[0].Content != "sys v2" || !strings.HasPrefix(last[len(last)-1].Content, refreshedNote) {
		t.Fatalf("after compaction: system %q, last %q", last[0].Content, last[len(last)-1].Content)
	}
	step(Request{Prompt: "three", Model: "m"})
	last = fake.sent(fake.count() - 1)
	if strings.Contains(last[len(last)-1].Content, "system-refreshed") {
		t.Errorf("a hot cache must not refresh: %q", last[len(last)-1].Content)
	}
}
