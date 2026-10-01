package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
)

func foldSession() *chatSession {
	s := &chatSession{width: 60}
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "question"})
	s.appendEntry(chatEntry{kind: core.UpdateReasoning, text: "thinking hard"})
	s.appendEntry(chatEntry{kind: core.UpdateToolCall, tool: "bash", text: "npm test"})
	s.appendEntry(chatEntry{kind: core.UpdateAssistant, text: "answer"})
	return s
}

func plain(rows []chatRow) string {
	var b strings.Builder
	for _, row := range rows {
		b.WriteString(row.text)
		for _, span := range row.spans {
			b.WriteString(span.text)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func TestFoldModesFilterTheTimeline(t *testing.T) {
	cases := []struct {
		mode      foldMode
		reasoning bool
		toolCall  bool
		nextHint  string
	}{
		{foldAll, true, true, "hide tool calls"},
		{foldNoTools, true, false, "hide reasoning too"},
		{foldMessages, false, false, "show everything"},
	}
	for _, c := range cases {
		s := foldSession()
		s.setFold(c.mode)
		text := plain(s.rows)
		if !strings.Contains(text, "question") || !strings.Contains(text, "answer") {
			t.Errorf("mode %d lost the messages:\n%s", c.mode, text)
		}
		if got := strings.Contains(text, "thinking hard"); got != c.reasoning {
			t.Errorf("mode %d reasoning shown = %v", c.mode, got)
		}
		if got := strings.Contains(text, "npm test"); got != c.toolCall {
			t.Errorf("mode %d tool call shown = %v", c.mode, got)
		}
		if !strings.Contains(c.mode.hint(), c.nextHint) {
			t.Errorf("mode %d hint = %q", c.mode, c.mode.hint())
		}
	}
}

func TestFoldIsReversible(t *testing.T) {
	s := foldSession()
	before := plain(s.rows)
	s.setFold(foldMessages)
	s.setFold(foldAll)
	if plain(s.rows) != before {
		t.Errorf("rows changed after a full cycle:\n%s\nvs\n%s", plain(s.rows), before)
	}
}

func TestFoldMode3FoldsStreamingEntries(t *testing.T) {
	s := foldSession()
	s.setFold(foldMessages)
	s.showUpdate(core.Update{Kind: core.UpdateReasoningDelta, Text: "more thoughts"})
	s.showUpdate(core.Update{Kind: core.UpdateAssistantDelta, Text: "next answer"})
	text := plain(s.rows)
	if strings.Contains(text, "more thoughts") || !strings.Contains(text, "next answer") {
		t.Errorf("streaming entries not folded:\n%s", text)
	}
	if len(s.history) != 6 {
		t.Errorf("hidden entries must stay in history, got %d", len(s.history))
	}
}

func TestTailShowsWorkingWithLastActionWhenFolded(t *testing.T) {
	a := &app{}
	s := foldSession()
	s.working = true
	s.setFold(foldNoTools)
	tail := plain(a.tailRows(s))
	for _, want := range []string{"working...", "bash", "npm test", "Ctrl+O to hide reasoning too"} {
		if !strings.Contains(tail, want) {
			t.Errorf("tail lacks %q:\n%s", want, tail)
		}
	}
	s.setFold(foldAll)
	if tail = plain(a.tailRows(s)); strings.Contains(tail, "working...") || !strings.Contains(tail, "Ctrl+O to hide tool calls") {
		t.Errorf("unfolded chat shows only the hint:\n%s", tail)
	}
}

func TestTailInStrictestFoldNamesReasoningAsLastAction(t *testing.T) {
	a := &app{}
	s := &chatSession{width: 100, working: true, fold: foldMessages}
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "question"})
	s.appendEntry(chatEntry{kind: core.UpdateReasoning, text: "weighing options"})
	if tail := plain(a.tailRows(s)); !strings.Contains(tail, "thinking") || !strings.Contains(tail, "weighing options") {
		t.Errorf("tail:\n%s", tail)
	}
}

func TestTailIsEmptyWithNothingToFold(t *testing.T) {
	s := &chatSession{width: 60}
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "hi"})
	if rows := (&app{}).tailRows(s); rows != nil {
		t.Errorf("unexpected tail: %v", rows)
	}
}

func TestCycleFoldAppliesToAllSessionsAndSaves(t *testing.T) {
	db, _ := openFoldDB(t)
	other := foldSession()
	a := &app{store: db, active: foldSession(), fold: foldAll}
	a.sessions = map[string]*chatSession{"a": a.active, "b": other}
	a.cycleFold()
	if a.fold != foldNoTools || other.fold != foldNoTools || a.active.fold != foldNoTools {
		t.Fatalf("fold not applied everywhere: %v %v %v", a.fold, other.fold, a.active.fold)
	}
	if cfg, _ := db.LoadConfig(); cfg.Fold != int(foldNoTools) {
		t.Errorf("saved fold = %d", cfg.Fold)
	}
	a.cycleFold()
	a.cycleFold()
	if a.fold != foldAll {
		t.Errorf("fold should wrap to all, got %d", a.fold)
	}
}

func TestDrawShowsWorkingLineAtTheEndOfTheChat(t *testing.T) {
	a, screen := layoutApp(t)
	a.active = foldSession()
	a.active.working = true
	a.active.setFold(foldNoTools)
	a.draw()
	found := false
	for y := 0; y < 24; y++ {
		if got := rowText(screen, y, 60); strings.Contains(got, "working...") && strings.Contains(got, "Ctrl+O") {
			found = true
		}
	}
	if !found {
		t.Error("working line with the Ctrl+O hint is not on screen")
	}
}
