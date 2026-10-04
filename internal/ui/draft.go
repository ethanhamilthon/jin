package ui

import (
	"jin/internal/core"
	"slices"
	"strings"
)

// cutRange removes input[start:end] and one neighbouring space, so no double
// or trailing space is left behind. The cursor lands where the cut was.
func cutRange(s *chatSession, start, end int) {
	switch {
	case start > 0 && end < len(s.input) && s.input[start-1] == " " && s.input[end] == " ":
		end++
	case start > 0 && end == len(s.input) && s.input[start-1] == " ":
		start--
	}
	s.input = slices.Delete(s.input, start, end)
	s.cursor = start
}

// inlineCommand finds the first command token anywhere in the draft. The
// arguments of a command that takes them run to the end of the line.
func inlineCommand(input []string) (cmd slashCommand, start, end int, arg string, ok bool) {
	for i, cluster := range input {
		t, isToken := tokenOf(cluster)
		if !isToken || t.kind != tokenCommand {
			continue
		}
		c, found := slashByName(t.payload)
		if !found {
			continue
		}
		end = i + 1
		for c.args && end < len(input) && input[end] != "\n" {
			end++
		}
		return c, i, end, strings.TrimSpace(draftPayload(input[i+1 : end])), true
	}
	return slashCommand{}, 0, 0, "", false
}

// runInlineCommand runs a command token such as "/tui lazygit" on Enter. It
// reports whether the draft held one, in which case it is not sent.
func (a *app) runInlineCommand() bool {
	s := a.active
	cmd, start, end, arg, ok := inlineCommand(s.input)
	if !ok {
		return false
	}
	if cmd.args && arg == "" {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "/" + cmd.name + " needs an argument: " + cmd.desc})
		return true
	}
	cutRange(s, start, end)
	cmd.run(a, arg)
	return true
}

// report shows an error from a command in the chat.
func (a *app) report(err error) {
	if err == nil {
		return
	}
	s := a.active
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateError, text: err.Error()})
}

func (a *app) clearDraft() {
	s := a.active
	s.input, s.cursor, s.inputTop = nil, 0, 0
	a.pruneTokens()
}

func (a *app) copyDraft() {
	copySelection(a.screen, draftPayload(a.active.input))
}

// newSessionWithDraft starts a session and moves the rest of the draft into it.
func (a *app) newSessionWithDraft() {
	old := a.active
	draft := slices.Clone(old.input)
	old.input, old.cursor, old.inputTop = nil, 0, 0
	a.newSession()
	a.active.input, a.active.cursor = draft, len(draft)
}
