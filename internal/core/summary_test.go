package core

import (
	"testing"

	"jin/internal/provider"
)

func TestNeedsCompaction(t *testing.T) {
	cases := []struct {
		size, window int
		want         bool
	}{{79, 100, false}, {80, 100, true}, {500, 0, false}, {0, 100, false}}
	for _, c := range cases {
		if got := needsCompaction(c.size, c.window); got != c.want {
			t.Errorf("needsCompaction(%d, %d) = %v", c.size, c.window, got)
		}
	}
}

func TestSinceLastSummary(t *testing.T) {
	user := func(text string) provider.Message { return provider.Message{Role: "user", Content: text} }
	messages := []provider.Message{user("a"), SummaryMessage("one"), user("b"), SummaryMessage("two"), user("c")}
	got := SinceLastSummary(messages)
	if len(got) != 2 || !IsSummary(got[0]) || got[1].Content != "c" {
		t.Fatalf("got %+v", got)
	}
	plain := []provider.Message{user("a"), user("b")}
	if len(SinceLastSummary(plain)) != 2 {
		t.Fatal("history without a summary must stay whole")
	}
}
