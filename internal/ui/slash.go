package ui

import (
	"slices"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

// slashCommand is one /command. A command with args takes the rest of the
// line, so accepting it from the list only completes its name.
type slashCommand struct {
	name string
	icon string
	desc string
	args bool
	run  func(a *app, arg string)
}

// slashCommands is the whole command list, in the order the list shows it.
func slashCommands() []slashCommand {
	return []slashCommand{
		{name: "sessions", icon: "☰", desc: "Sessions of this directory", run: func(a *app, _ string) { a.openTab(tabSessions) }},
		{name: "prompts", icon: "✎", desc: "Reusable prompts: edit, add, delete", run: func(a *app, _ string) { a.openTab(tabPrompts) }},
		{name: "hooks", icon: "◈", desc: "Hooks: edit, add, delete, on/off", run: func(a *app, _ string) { a.openTab(tabHooks) }},
		{name: "model", icon: "◆", desc: "Select the model", run: func(a *app, _ string) { a.openModelFlow() }},
		{name: "scope", icon: "◇", desc: "Choose the models Ctrl+M cycles through", run: func(a *app, _ string) { a.openScopeFlow() }},
		{name: "provider", icon: "⇄", desc: "Providers: add, switch, delete", run: func(a *app, _ string) { a.openProviderFlow() }},
		{name: "sound", icon: "♪", desc: "Notification sound", run: func(a *app, _ string) { a.openSoundFlow() }},
		{name: "tools", icon: "⚒", desc: "Switch agent tools on and off", run: func(a *app, _ string) { a.openToolsFlow() }},
		{name: "change-editor", icon: "✐", desc: "Choose the external editor", run: func(a *app, _ string) {
			a.chooseEditor(func() error { return nil })
		}},
		{name: "compact", icon: "≋", desc: "Summarize the conversation to free context", run: func(a *app, _ string) { a.compactSession() }},
		{name: "handoff", icon: "➜", desc: "Continue the work in a fresh session", run: func(a *app, _ string) { a.handoffSession() }},
		{name: "rewind", icon: "⟲", desc: "Start over from one of your messages in a new session", run: func(a *app, _ string) { a.openRewindFlow() }},
		{name: "undo", icon: "↶", desc: "Undo the file changes of the agent's last turn", run: func(a *app, _ string) { a.undoLastTurn() }},
		{name: "stop", icon: "■", desc: "Interrupt the running request", run: func(a *app, _ string) { a.active.agent.Interrupt() }},
		{name: "new", icon: "+", desc: "New session, the draft moves into it", run: func(a *app, _ string) { a.newSessionWithDraft() }},
		{name: "quit", icon: "✕", desc: "Quit jin", run: func(a *app, _ string) { a.requestQuit() }},
		{name: "clear", icon: "⌫", desc: "Clear the draft", run: func(a *app, _ string) { a.clearDraft() }},
		{name: "copy", icon: "⎘", desc: "Copy the draft to the clipboard", run: func(a *app, _ string) { a.copyDraft() }},
		{name: "edit", icon: "✎", desc: "Edit the draft in the editor", run: func(a *app, _ string) { a.report(a.editInput(a.active)) }},
		{name: "todo", icon: "☑", desc: "Edit the todo list in the editor", run: func(a *app, _ string) { a.editTodos(a.active) }},
		{name: "tui", icon: "▣", desc: "Run a full-screen program: /tui lazygit", args: true, run: func(a *app, arg string) { a.runTUI(arg) }},
		{name: "async-tasks", icon: "◐", desc: "Background tasks: output and stop", run: func(a *app, _ string) { a.openAsyncTasksFlow() }},
		{name: "system-prompt", icon: "§", desc: "Edit the system, compact and handoff prompts", run: func(a *app, _ string) { a.openSystemPromptFlow() }},
		{name: "bash", icon: "$", desc: "Shell input until Esc; output stays out of the chat context", run: func(a *app, _ string) { a.startBash() }},
	}
}

func slashByName(name string) (slashCommand, bool) {
	for _, c := range slashCommands() {
		if c.name == name {
			return c, true
		}
	}
	return slashCommand{}, false
}

// slash is the autocomplete panel for a /command typed anywhere in the input.
type slash struct {
	sel   *selector
	start int
}

// closedToken remembers an autocomplete closed with Esc, so the token stays
// plain text until it changes. kind is the trigger: '/', '@' or '#'.
type closedToken struct {
	session string
	kind    byte
	start   int
	query   string
}

// forgetClosed drops the Esc memory once its token is gone from the input,
// so the same trigger typed again later opens its list.
func (a *app) forgetClosed(kind byte, start int, found bool) {
	d := a.closed
	if d.kind == kind && (!found || d.session != a.active.id || d.start != start) {
		a.closed = closedToken{}
	}
}

func isSlashNameCluster(cluster string) bool {
	runes := []rune(cluster)
	return len(runes) == 1 && (unicode.IsLetter(runes[0]) || unicode.IsDigit(runes[0]) || runes[0] == '-')
}

// slashAt finds the /token that ends at the cursor. The / must start the text
// or follow whitespace, so paths and "and/or" are never commands.
func slashAt(input []string, cursor int) (start int, query []string, ok bool) {
	for i := cursor - 1; i >= 0; i-- {
		switch cluster := input[i]; {
		case cluster == "/":
			if i == 0 || strings.TrimSpace(input[i-1]) == "" {
				return i, input[i+1 : cursor], true
			}
			return 0, nil, false
		case !isSlashNameCluster(cluster):
			return 0, nil, false
		}
	}
	return 0, nil, false
}

func (a *app) refreshSlash() {
	previous, previousStart := "", -1
	if a.slash != nil {
		previous, previousStart = a.slash.sel.current(), a.slash.start
	}
	a.slash = nil
	s := a.active
	if a.sel != nil || s.ask != nil || s.bash != nil {
		return
	}
	start, query, ok := slashAt(s.input, s.cursor)
	a.forgetClosed('/', start, ok)
	if !ok {
		return
	}
	typed := strings.ToLower(strings.Join(query, ""))
	if d := a.closed; d.session == s.id && d.kind == '/' && d.start == start && d.query == typed {
		return
	}
	var options []option
	for _, c := range slashCommands() {
		if strings.HasPrefix(c.name, typed) {
			options = append(options, option{label: "/" + c.name, detail: c.desc, value: c.name})
		}
	}
	if len(options) == 0 {
		return
	}
	icons := map[string]string{}
	for _, c := range slashCommands() {
		icons[c.name] = c.icon
	}
	sel := &selector{title: "Commands", options: options, mark: func(name string) string { return icons[name] }}
	if start == previousStart {
		sel.selectValue(previous)
	}
	a.slash = &slash{sel: sel, start: start}
}

// slashKey lets the open list take navigation, completion and Esc. Enter
// completes only while the list is open, so a typed /name never sends.
func (a *app) slashKey(ev *tcell.EventKey) bool {
	p := a.slash
	if p == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyUp:
		p.sel.move(-1)
	case tcell.KeyDown:
		p.sel.move(1)
	case tcell.KeyTab:
		a.acceptSlash()
	case tcell.KeyEnter:
		if a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0 {
			return false
		}
		a.acceptSlash()
	case tcell.KeyEscape:
		s := a.active
		a.closed = closedToken{session: s.id, kind: '/', start: p.start, query: strings.ToLower(strings.Join(s.input[p.start+1:s.cursor], ""))}
		a.slash = nil
	default:
		return false
	}
	return true
}

// acceptSlash runs the highlighted command and cuts its token out of the
// draft. A command that takes arguments only gets its name completed.
func (a *app) acceptSlash() {
	s, p := a.active, a.slash
	cmd, ok := slashByName(p.sel.current())
	if !ok {
		return
	}
	if cmd.args {
		completion := clusters(cmd.name + " ")
		tail := append(completion, s.input[s.cursor:]...)
		s.input = append(s.input[:p.start+1], tail...)
		s.cursor = p.start + 1 + len(completion)
		return
	}
	cutRange(s, p.start, s.cursor)
	a.slash = nil
	cmd.run(a, "")
}

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

// inlineCommand finds "/name args" for a command that takes arguments, anywhere
// in the draft. The arguments run to the end of the line.
func inlineCommand(input []string) (cmd slashCommand, start, end int, arg string, ok bool) {
	for i, cluster := range input {
		if cluster != "/" || (i > 0 && strings.TrimSpace(input[i-1]) != "") {
			continue
		}
		j := i + 1
		for j < len(input) && isSlashNameCluster(input[j]) {
			j++
		}
		c, found := slashByName(strings.Join(input[i+1:j], ""))
		if !found || !c.args || j >= len(input) || input[j] != " " {
			continue
		}
		end = j
		for end < len(input) && input[end] != "\n" {
			end++
		}
		return c, i, end, strings.TrimSpace(strings.Join(input[j:end], "")), true
	}
	return slashCommand{}, 0, 0, "", false
}

// runInlineCommand runs a typed "/tui lazygit" on Enter. It reports whether
// the draft held one, in which case it is not sent.
func (a *app) runInlineCommand() bool {
	s := a.active
	cmd, start, end, arg, ok := inlineCommand(s.input)
	if !ok {
		return false
	}
	if arg == "" {
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
}

func (a *app) copyDraft() {
	copySelection(a.screen, strings.Join(a.active.input, ""))
}

// newSessionWithDraft starts a session and moves the rest of the draft into it.
func (a *app) newSessionWithDraft() {
	old := a.active
	draft := slices.Clone(old.input)
	old.input, old.cursor, old.inputTop = nil, 0, 0
	a.newSession()
	a.active.input, a.active.cursor = draft, len(draft)
}
