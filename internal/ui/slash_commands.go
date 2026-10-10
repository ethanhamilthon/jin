package ui

import "strings"

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
// The global settings live behind /settings.
func slashCommands() []slashCommand {
	return []slashCommand{
		{name: "sessions", icon: "☰", desc: "Sessions of this project: resume, search, new", run: func(a *app, _ string) { a.openProjectSessions(a.dir) }},
		{name: "projects", icon: "❖", desc: "Projects: sessions, add, rename, archive", run: func(a *app, _ string) { a.openProjects() }},
		{name: "model", icon: "◆", desc: "Select the model", run: func(a *app, _ string) { a.openModelFlow() }},
		{name: "provider", icon: "⇄", desc: "Providers: add, switch, delete", run: func(a *app, _ string) { a.openProviderFlow() }},
		{name: "theme", icon: "◑", desc: "Color theme", run: func(a *app, _ string) { a.openThemeFlow() }},
		{name: "settings", icon: "⚙", desc: "Sound, tools, prompts, hooks, data folder and more", run: func(a *app, _ string) { a.openSettingsFlow() }},
		{name: "reload", icon: "↻", desc: "Reload this session's prompts, hooks and instructions", run: func(a *app, _ string) { a.reloadSession() }},
		{name: "compact", icon: "≋", desc: "Summarize the conversation to free context", run: func(a *app, _ string) { a.compactSession() }},
		{name: "title", icon: "❝", desc: "Name this session now with the title model", run: func(a *app, _ string) { a.nameNow() }},
		{name: "export", icon: "⤓", desc: "Save this session as Markdown in the project folder", run: func(a *app, _ string) { a.exportSession() }},
		{name: "context", icon: "◫", desc: "Show what fills the context; /context full prints the text of every part", args: true, run: func(a *app, arg string) { a.showContext(strings.TrimSpace(arg) == "full") }},
		{name: "handoff", icon: "➜", desc: "Continue the work in a fresh session", run: func(a *app, _ string) { a.handoffSession() }},
		{name: "rewind", icon: "⟲", desc: "Restart the conversation from a message; it does not change files", run: func(a *app, _ string) { a.openRewindFlow() }},
		{name: "resume", icon: "▶", desc: "Continue the shared queue after Stop", run: func(a *app, _ string) {
			if a.backend != nil {
				a.backendAction("resume")
			}
		}},
		{name: "stop", icon: "■", desc: "Interrupt the focused shell command or request", run: func(a *app, _ string) { a.interrupt() }},
		{name: "new", icon: "+", desc: "New session, the draft moves into it", run: func(a *app, _ string) { a.newSessionWithDraft() }},
		{name: "clear", icon: "⌫", desc: "Clear the draft", run: func(a *app, _ string) { a.clearDraft() }},
		{name: "edit", icon: "✎", desc: "Edit the draft in the editor", run: func(a *app, _ string) { a.report(a.editInput(a.active)) }},
		{name: "tui", icon: "▣", desc: "Run a full-screen program: /tui lazygit", args: true, run: func(a *app, arg string) { a.runTUI(arg) }},
		{name: "vertical", icon: "▥", desc: "Split the focused pane left and right", run: func(a *app, _ string) { a.splitPane(true) }},
		{name: "horizontal", icon: "▤", desc: "Split the focused pane top and bottom", run: func(a *app, _ string) { a.splitPane(false) }},
		{name: "quit", icon: "✕", desc: "Close the focused pane", run: func(a *app, _ string) { a.closePane() }},
		{name: "qa", icon: "⏻", desc: "Quit Jin", run: func(a *app, _ string) { a.requestQuit() }},
		{name: "tasks", icon: "◐", desc: "Background tasks: output and stop", run: func(a *app, _ string) { a.openTasksFlow() }},
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
