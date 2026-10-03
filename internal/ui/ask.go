package ui

import (
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
