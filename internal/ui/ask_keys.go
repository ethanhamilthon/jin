package ui

import (
	"jin/internal/core"
	"jin/internal/daemon"
	"jin/internal/tools"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
)

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
	if a.backend != nil {
		err := a.backend.Command(a.ctx, daemon.Command{Action: "answer", Session: s.id, Question: s.remoteState.Question, Answers: q.answers}, nil)
		a.backendError(err)
		s.ask = nil
		return
	}
	s.ask = nil
	s.appendEntry(chatEntry{kind: core.UpdateAsk, text: "Questions\n" + tools.FormatAnswers(q.questions, q.answers)})
	s.scroll = 0
	s.agent.Answer(q.answers)
}
