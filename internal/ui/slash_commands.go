package ui

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
		{name: "theme", icon: "◑", desc: "Color theme", run: func(a *app, _ string) { a.openThemeFlow() }},
		{name: "motion", icon: "≈", desc: "Animation speed of the glow, or off", run: func(a *app, _ string) { a.openMotionFlow() }},
		{name: "sound", icon: "♪", desc: "Notification sound", run: func(a *app, _ string) { a.openSoundFlow() }},
		{name: "tools", icon: "⚒", desc: "Switch agent tools on and off", run: func(a *app, _ string) { a.openToolsFlow() }},
		{name: "change-editor", icon: "✐", desc: "Choose the external editor", run: func(a *app, _ string) {
			a.chooseEditor(func() error { return nil })
		}},
		{name: "compact", icon: "≋", desc: "Summarize the conversation to free context", run: func(a *app, _ string) { a.compactSession() }},
		{name: "handoff", icon: "➜", desc: "Continue the work in a fresh session", run: func(a *app, _ string) { a.handoffSession() }},
		{name: "rewind", icon: "⟲", desc: "Restart the conversation from a message; it does not change files", run: func(a *app, _ string) { a.openRewindFlow() }},
		{name: "undo", icon: "↶", desc: "Restore edit and write changes of the last turn; bash changes are not covered", run: func(a *app, _ string) { a.undoLastTurn() }},
		{name: "diff", icon: "±", desc: "Diff of the edit and write changes of the last turn", args: true, run: func(a *app, arg string) { a.showDiff(arg) }},
		{name: "stop", icon: "■", desc: "Interrupt the running request", run: func(a *app, _ string) { a.active.agent.Interrupt() }},
		{name: "new", icon: "+", desc: "New session, the draft moves into it", run: func(a *app, _ string) { a.newSessionWithDraft() }},
		{name: "quit", icon: "✕", desc: "Quit jin", run: func(a *app, _ string) { a.requestQuit() }},
		{name: "clear", icon: "⌫", desc: "Clear the draft", run: func(a *app, _ string) { a.clearDraft() }},
		{name: "links", icon: "↗", desc: "Open a link of the last answer", run: func(a *app, _ string) { a.openLinksFlow() }},
		{name: "copy", icon: "⎘", desc: "Copy the draft to the clipboard", run: func(a *app, _ string) { a.copyDraft() }},
		{name: "edit", icon: "✎", desc: "Edit the draft in the editor", run: func(a *app, _ string) { a.report(a.editInput(a.active)) }},
		{name: "todo", icon: "☑", desc: "Edit the todo list in the editor", run: func(a *app, _ string) { a.editTodos(a.active) }},
		{name: "tui", icon: "▣", desc: "Run a full-screen program: /tui lazygit", args: true, run: func(a *app, arg string) { a.runTUI(arg) }},
		{name: "async-tasks", icon: "◐", desc: "Background tasks: output and stop", run: func(a *app, _ string) { a.openAsyncTasksFlow() }},
		{name: "system-prompt", icon: "§", desc: "Edit the system, compact and handoff prompts", run: func(a *app, _ string) { a.openSystemPromptFlow() }},
		{name: "reset", icon: "⟳", desc: "Move all jin data aside and start from scratch", run: func(a *app, _ string) { a.openResetFlow() }},
		{name: "swap-config", icon: "⇆", desc: "Use another jin data folder instead of the current one", run: func(a *app, _ string) { a.openSwapFlow() }},
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
