package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
	"jin/internal/tools"
)

func askApp(t *testing.T, questions ...tools.Question) *app {
	t.Helper()
	a, _ := layoutApp(t)
	a.active.ask = newAskState(questions)
	a.active.agent = core.NewAgent(nil, "", tools.NewRegistry())
	return a
}

func TestAskLeftRightMoveBetweenQuestionsAndKeepAnswers(t *testing.T) {
	a := askApp(t,
		tools.Question{Question: "one", Options: []string{"a", "b"}},
		tools.Question{Question: "two", Options: []string{"c", "d"}},
		tools.Question{Question: "three", Options: []string{"e", "f"}})
	q := a.active.ask
	press(a, tcell.KeyEnter) // answer "a"
	if q.current != 1 {
		t.Fatalf("current = %d", q.current)
	}
	press(a, tcell.KeyLeft)
	if q.current != 0 || q.answers[0] != "a" {
		t.Fatalf("Left should go back and keep the answer: current=%d answers=%v", q.current, q.answers)
	}
	press(a, tcell.KeyRight)
	if q.current != 1 {
		t.Fatalf("Right should go forward to an answered neighbour, current = %d", q.current)
	}
	press(a, tcell.KeyRight)
	if q.current != 1 {
		t.Fatal("Right must not skip an unanswered question")
	}
}

func TestAskOverwritesAnEarlierAnswer(t *testing.T) {
	a := askApp(t,
		tools.Question{Question: "one", Options: []string{"a", "b"}},
		tools.Question{Question: "two", Options: []string{"c", "d"}})
	q := a.active.ask
	press(a, tcell.KeyEnter)
	press(a, tcell.KeyLeft)
	press(a, tcell.KeyDown)
	press(a, tcell.KeyEnter) // overwrite with "b", move on
	if q.answers[0] != "b" || len(q.answers) != 1 || q.current != 1 {
		t.Fatalf("answers=%v current=%d", q.answers, q.current)
	}
}

func TestAskMultiSelectToggleAndSend(t *testing.T) {
	a := askApp(t, tools.Question{Question: "pick", Options: []string{"x", "y", "z"}, Multiple: true})
	typeText(a, " ") // toggle x
	press(a, tcell.KeyDown)
	press(a, tcell.KeyDown)
	typeText(a, " ") // toggle z
	typeText(a, " ") // untoggle z
	typeText(a, " ") // toggle z again
	line := strings.Join(rowsOf(a.active.ask), "\n")
	if !strings.Contains(line, "[x] 1. x") || !strings.Contains(line, "[x] 3. z") || !strings.Contains(line, "[ ] 2. y") {
		t.Fatalf("rows:\n%s", line)
	}
	q := a.active.ask
	press(a, tcell.KeyEnter)
	if a.active.ask != nil {
		t.Fatal("answering the last question should close the ask block")
	}
	_ = q
}

func TestAskSpaceDoesNothingOnASingleChoiceQuestion(t *testing.T) {
	a := askApp(t, tools.Question{Question: "pick", Options: []string{"x", "y"}})
	typeText(a, " ")
	if strings.Contains(strings.Join(rowsOf(a.active.ask), "\n"), "[x]") {
		t.Fatal("Space must not select on a single-choice question")
	}
}

func TestAskFreeTextRowKeepsCursorKeys(t *testing.T) {
	a := askApp(t, tools.Question{Question: "q", Options: []string{"x"}}, tools.Question{Question: "r"})
	press(a, tcell.KeyDown) // free-text row
	typeText(a, "ab")
	press(a, tcell.KeyLeft)
	if a.active.ask.current != 0 || a.active.ask.cursor != 1 {
		t.Fatalf("Left on typed text must move the text cursor: current=%d cursor=%d", a.active.ask.current, a.active.ask.cursor)
	}
}

func TestFormatAnswersBulletsForSeveralOptions(t *testing.T) {
	qs := []tools.Question{{Question: "pick", Multiple: true}, {Question: "one"}}
	got := tools.FormatAnswers(qs, []string{"- x, with comma\n- y", "z"})
	want := "pick →\n- x, with comma\n- y\none → z"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestAskSchemaHasMultiple(t *testing.T) {
	for _, s := range tools.Catalog() {
		_ = s
	}
	if !strings.Contains(string(tools.Build([]string{"ask_user"}).SchemaJSON()), `"multiple"`) {
		t.Fatal("ask_user schema should offer the multiple field")
	}
}

func rowsOf(q *askState) []string {
	rows, _ := q.lines(60)
	texts := make([]string, len(rows))
	for i, row := range rows {
		texts[i] = row.text
	}
	return texts
}
