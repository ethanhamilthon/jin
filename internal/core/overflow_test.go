package core

import (
	"strings"
	"testing"

	"jin/internal/provider"
)

func TestTurnRecoversFromOverflowByCompacting(t *testing.T) {
	agent, fake := newFakeAgent(t, "overflow", "the summary", "answer")
	history := toolTurns(6, 400)
	updates := make(chan Update, 256)
	agent.turn(t.Context(), Request{Prompt: "next", Model: "m"}, &history, nil, updates)
	all := collect(updates)
	if fake.count() != 3 {
		t.Fatalf("requests = %d, want 3", fake.count())
	}
	if n := omitted(fake.sent(1)); n != 3 {
		t.Errorf("compaction request omitted %d results, want 3", n)
	}
	if retry := fake.sent(2); len(retry) != 2 || !IsSummary(retry[1]) {
		t.Errorf("retry sent %+v, want system and summary", retry)
	}
	var notices int
	var done Update
	for _, u := range all {
		if u.Kind == UpdateInfo && strings.Contains(u.Text, "context window is full") {
			notices++
		}
		if u.Kind == UpdateDone {
			done = u
		}
	}
	if notices != 1 || !done.Final {
		t.Errorf("notices = %d, done = %+v", notices, done)
	}
}

func TestTurnRecoversFromOverflowByPruning(t *testing.T) {
	agent, fake := newFakeAgent(t, "overflow", "answer")
	history := append(toolTurns(6, 2000), provider.Message{Role: "user", Content: "next"})
	updates := make(chan Update, 256)
	err := agent.answer(t.Context(), t.Context(), Request{Model: "m", Window: 2400}, &history, nil, updates)
	all := collect(updates)
	if err != nil || fake.count() != 2 {
		t.Fatalf("err = %v, requests = %d", err, fake.count())
	}
	if n := omitted(fake.sent(1)); n != 3 {
		t.Errorf("retry omitted %d results, want 3", n)
	}
	for _, u := range all {
		if u.Kind == UpdateCompacted {
			t.Error("compacted although pruning was enough")
		}
	}
}

func TestCompactionRecoversFromOverflow(t *testing.T) {
	agent, fake := newFakeAgent(t, "overflow", "the summary")
	history := toolTurns(4, 400)
	updates := make(chan Update, 256)
	if err := agent.compact(t.Context(), t.Context(), Request{Model: "m"}, &history, updates); err != nil {
		t.Fatal(err)
	}
	collect(updates)
	trimmed := fake.sent(1)
	if n := omitted(trimmed); n != 2 {
		t.Errorf("trimmed request omitted %d results, want 2", n)
	}
	if last := trimmed[len(trimmed)-3]; last.Role != "tool" || len(last.Content) != 400 {
		t.Errorf("last turn was trimmed: %+v", last)
	}
	if len(history) != 2 || !IsSummary(history[1]) {
		t.Errorf("history = %+v", history)
	}
}

func TestSecondOverflowIsReported(t *testing.T) {
	cases := []struct {
		name    string
		replies []string
	}{
		{"compaction overflows too", []string{"overflow"}},
		{"retry overflows again", []string{"overflow", "the summary", "overflow"}},
	}
	for _, c := range cases {
		agent, fake := newFakeAgent(t, c.replies...)
		history := toolTurns(6, 400)
		updates := make(chan Update, 256)
		agent.turn(t.Context(), Request{Prompt: "next", Model: "m"}, &history, nil, updates)
		var failed bool
		for _, u := range collect(updates) {
			failed = failed || u.Kind == UpdateError
			if u.Kind == UpdateDone && u.Final {
				t.Errorf("%s: turn reported success", c.name)
			}
		}
		if !failed || fake.count() != 3 {
			t.Errorf("%s: failed = %v, requests = %d, want 3", c.name, failed, fake.count())
		}
	}
}
