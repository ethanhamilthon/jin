package ui

import (
	"strconv"
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
	"jin/internal/tools"
)

// maxBlockRows is the tallest the ask and todo blocks get; longer content
// scrolls.
const maxBlockRows = 7

// askState is a pending ask_user call of one session. Every question has its
// options plus a free-text row at the end.
type askState struct {
	questions []tools.Question
	current   int
	row       int // selected row; len(options) is the free-text row
	text      []string
	cursor    int
	answers   []string
	selected  [][]bool
	top       int
}

func newAskState(questions []tools.Question) *askState {
	return &askState{questions: questions}
}

func (q *askState) question() tools.Question { return q.questions[q.current] }

func (q *askState) onFreeRow() bool { return q.row >= len(q.question().Options) }

func (q *askState) toggle(question, option int) {
	for len(q.selected) <= question {
		q.selected = append(q.selected, nil)
	}
	for len(q.selected[question]) < len(q.questions[question].Options) {
		q.selected[question] = append(q.selected[question], false)
	}
	q.selected[question][option] = !q.selected[question][option]
}

// askKey handles a key while the ask block is on screen and no panel is open.
func (a *app) askKey(ev *tcell.EventKey) {
	s := a.active
	q := s.ask
	switch {
	case ev.Key() == tcell.KeyUp:
		q.row = max(0, q.row-1)
	case ev.Key() == tcell.KeyDown:
		q.row = min(len(q.question().Options), q.row+1)
	case ev.Key() == tcell.KeyLeft && (!q.onFreeRow() || (q.cursor == 0 && len(q.text) == 0)):
		if q.current > 0 {
			q.current--
			q.row, q.text, q.cursor, q.top = 0, nil, 0, 0
		}
	case ev.Key() == tcell.KeyRight && !q.onFreeRow():
		if q.current+1 < len(q.questions) && q.current < len(q.answers) {
			q.current++
			q.row, q.text, q.cursor, q.top = 0, nil, 0, 0
		}
	case ev.Key() == tcell.KeyRune && ev.Str() == " " && !q.onFreeRow() && q.question().Multiple:
		q.toggle(q.current, q.row)
	case isPasteKey(ev) && q.onFreeRow():
		if text, ok := pasteClipboard(); ok {
			insertClusters(&q.text, &q.cursor, strings.ReplaceAll(text, "\n", " "))
		}
	case ev.Key() == tcell.KeyEnter && !q.onFreeRow():
		answer := q.question().Options[q.row]
		if q.question().Multiple {
			var picks []string
			if q.current < len(q.selected) {
				for i, yes := range q.selected[q.current] {
					if yes {
						picks = append(picks, q.question().Options[i])
					}
				}
			}
			if len(picks) > 0 {
				answer = "- " + strings.Join(picks, "\n- ")
			}
		}
		a.answerAsk(s, answer)
	case !q.onFreeRow():
		if ev.Key() == tcell.KeyRune && ev.Modifiers() == 0 {
			if n, err := strconv.Atoi(ev.Str()); err == nil && n >= 1 && n <= len(q.question().Options) {
				q.row = n - 1
			}
		}
	default:
		if text := handleInput(ev, &q.text, &q.cursor); text != "" {
			a.answerAsk(s, text)
		}
	}
}

// answerAsk records one answer and, after the last question, hands all of
// them to the agent and puts them in the timeline.
func (a *app) answerAsk(s *chatSession, answer string) {
	q := s.ask
	if q.current < len(q.answers) {
		q.answers[q.current] = answer
	} else {
		q.answers = append(q.answers, answer)
	}
	if q.current+1 < len(q.questions) {
		q.current++
		q.row, q.text, q.cursor, q.top = 0, nil, 0, 0
		return
	}
	s.ask = nil
	s.appendEntry(chatEntry{kind: core.UpdateAsk, text: "Questions\n" + tools.FormatAnswers(q.questions, q.answers)})
	s.scroll = 0
	s.agent.Answer(q.answers)
}

// askLines lays the current question out as rows and returns the row that
// holds the selection.
func (q *askState) lines(width int) (rows []string, selected int) {
	title := q.question().Question
	if len(q.questions) > 1 {
		title = "(" + strconv.Itoa(q.current+1) + "/" + strconv.Itoa(len(q.questions)) + ") " + title
	}
	rows = append(rows, wrapChat(title, width-2)...)
	for i, option := range q.question().Options {
		if i == q.row {
			selected = len(rows)
		}
		marker := "  "
		if i == q.row {
			marker = "› "
		}
		if q.question().Multiple {
			checked := q.current < len(q.selected) && i < len(q.selected[q.current]) && q.selected[q.current][i]
			box := "[ ] "
			if checked {
				box = "[x] "
			}
			marker += box
		}
		rows = append(rows, marker+strconv.Itoa(i+1)+". "+option)
	}
	if q.onFreeRow() {
		selected = len(rows)
	}
	rows = append(rows, "")
	return rows, selected
}

// askHeight is how many rows the block takes: the question and options, then
// the free-text row.
func (q *askState) height(width int) int {
	rows, _ := q.lines(width)
	return min(maxBlockRows, len(rows))
}

func (a *app) drawAsk(q *askState, top, height, w int) {
	a.screen.HideCursor()
	rows, selected := q.lines(w)
	rows = rows[:len(rows)-1] // the free-text row is drawn separately
	questionRows := height - 1
	q.top = fitScroll(q.top, min(selected, len(rows)-1), len(rows), questionRows)
	for i := 0; i < questionRows && q.top+i < len(rows); i++ {
		style := base
		if strings.HasPrefix(rows[q.top+i], "› ") {
			style = accent.Bold(true)
		}
		put(a.screen, 2, top+i, truncate(rows[q.top+i], w-3), style)
	}
	y := top + height - 1
	prefix, style := "  ", dim
	if q.onFreeRow() {
		prefix, style = "› ", accent.Bold(true)
	}
	put(a.screen, 2, y, prefix, style)
	if len(q.text) == 0 {
		put(a.screen, 4, y, "Type your own answer...", dim)
		if q.onFreeRow() {
			a.screen.ShowCursor(4, y)
		}
		return
	}
	line := strings.Join(q.text, "")
	put(a.screen, 4, y, truncate(strings.ReplaceAll(line, "\n", " "), w-5), base)
	if q.onFreeRow() {
		a.screen.ShowCursor(min(w-1, 4+displaywidth.String(strings.Join(q.text[:q.cursor], ""))), y)
	}
}
