package ui

import (
	"strconv"
	"strings"

	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
)

// rewindPoint is a message the user typed and where it sits in the history.
type rewindPoint struct {
	index int
	text  string
}

// typedText is what the user typed in a stored user message, without the
// blocks jin adds around it; ok is false for messages the user did not type.
func typedText(msg provider.Message) (string, bool) {
	if msg.Role != "user" || core.IsSummary(msg) || core.ImageLabels(msg) != nil ||
		strings.HasPrefix(msg.Content, "<task-result ") {
		return "", false
	}
	text := prompts.Strip(core.StripNotes(msg.Content))
	if i := strings.LastIndex(text, "\n\n<attached-files>\n"); i >= 0 && strings.HasSuffix(text, "</attached-files>") {
		text = text[:i]
	}
	return text, strings.TrimSpace(text) != ""
}

func rewindPoints(messages []provider.Message) []rewindPoint {
	var points []rewindPoint
	for i, msg := range messages {
		if text, ok := typedText(msg); ok {
			points = append(points, rewindPoint{index: i, text: text})
		}
	}
	return points
}

// openRewindFlow lists the user's messages, newest first. Choosing one
// starts a new session with the history before it and its text as the draft.
func (a *app) openRewindFlow() {
	s := a.active
	if !s.persisted {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to rewind: this session has no messages"})
		return
	}
	messages, err := a.store.LoadMessages(s.id)
	if err != nil {
		s.persistenceError("Rewind failed", err)
		return
	}
	points := rewindPoints(messages)
	var options []option
	for i := len(points) - 1; i >= 0; i-- {
		p := points[i]
		options = append(options, option{label: truncate(firstLine(p.text), 70), detail: "#" + strconv.Itoa(i+1), value: strconv.Itoa(i)})
	}
	if len(options) == 0 {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to rewind"})
		return
	}
	a.openList("Rewind · restarts the conversation from a message; it does not change files", options, options[0].value, func(value string) error {
		n, _ := strconv.Atoi(value)
		return a.forkAt(s, messages[:points[n].index], points[n].text)
	})
}

// forkAt saves a new session with history and opens it with draft as input.
// The original session is left as it is.
func (a *app) forkAt(from *chatSession, history []provider.Message, draft string) error {
	id := newSessionID()
	if err := a.store.TouchProvider(id, from.path, from.model, from.effort, from.title, from.provider); err != nil {
		return err
	}
	for _, msg := range history {
		if err := a.store.AppendMessage(id, msg); err != nil {
			return err
		}
	}
	rec, _, err := a.store.GetSession(id)
	if err != nil {
		return err
	}
	if err := a.resumeSession(rec); err != nil {
		return err
	}
	a.active.input = clusters(draft)
	a.active.cursor = len(a.active.input)
	a.sel = nil
	return nil
}
