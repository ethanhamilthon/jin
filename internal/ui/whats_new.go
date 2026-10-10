package ui

import "strings"

const keyReleaseSeen = "release.seen"

// releaseNotes are the highlights of each version, shown once in the intro
// after an upgrade. Keep each to a few short lines.
var releaseNotes = map[string]string{
	"v0.14.0": strings.Join([]string{
		"Providers have enable switches and models are picked as provider/model; subscriptions can run through a managed CLIProxyAPI",
		"Sessions are named by a model after your first message (/settings, Session titles); /title, /export and a session picker by project in the TUI and jin web",
		"Gone: /undo and removing projects (archive them instead); sessions of any age are listed",
	}, "\n"),
	"v0.13.0": strings.Join([]string{
		"The system prompt is now exactly your system-prompt.md with its commands run; /context full shows the text of every part",
		"New grep tool, jin docs (the docs inside the binary) and jin hooks render; hooks and AGENTS.md come in through commands",
		"Gone: the macOS sed -i and TODO.md lines, parent and nested AGENTS.md; the gallery has hooks for them",
	}, "\n"),
	"v0.12.0": strings.Join([]string{
		"jin web --remote reaches jin from your phone through Tailscale; scan the QR code behind the badge at the top",
		"jin web: Settings, Devices lists the browsers that can open jin and revokes them",
		"jin web: a phone-size window gets one pane, a sidebar drawer and bottom sheets",
	}, "\n"),
	"v0.11.0": strings.Join([]string{
		"jin web: Settings, Project and Files are panes next to your chats, and the layout is remembered",
		"jin web: archive projects from the sidebar and restore them in Settings; browse project files and insert @path",
		"jin web: new Markdown look, accent selection, and Chat details hides reasoning together with tool output",
	}, "\n"),
	"v0.10.0": strings.Join([]string{
		"The todo and tell_user tools are gone: the agent keeps its plan in TODO.md, and #plan answers with a numbered list",
		"jin sessions compact|handoff|rewind|undo|context|reload <id> work from scripts; jin web resets the system prompt to the latest from git",
		"jin web: selected text uses the accent color. On macOS the agent is told never to run sed -i",
	}, "\n"),
	"v0.9.6": strings.Join([]string{
		"Projects show as paths (~/code/app) in the TUI and jin web, not as names",
		"jin web: pane headers show the project path, the active pane follows the cursor, the ... menu has icons",
		"jin web: the chat stays where you read while new messages arrive; Rewind of an empty session no longer hangs",
	}, "\n"),
	"v0.9.5": strings.Join([]string{
		"The pinned todo list folds: Ctrl+T or a click on its line in the TUI, a click on [ todo ] in jin web",
		"jin web: accent border for the focused pane, a pulsing glow while working, violet badge and glow for tasks",
		"jin web: the suggestion shows in the input; Space edits it, Enter sends it. #subagents moved to the gallery",
	}, "\n"),
	"v0.9.4": strings.Join([]string{
		"Documentation patch: the README links to the jin web guide",
		"About explains the shared terminal and browser data; database docs cover web.accent",
		"All v0.9.3 features are included unchanged",
	}, "\n"),
	"v0.9.3": strings.Join([]string{
		"Pictures from read show in the chat, and Markdown images open on click, in the TUI and in jin web",
		"jin web: a paperclip attaches files, three dots hold compact, handoff, rewind and more; no / commands",
		"jin web: sessions sit under their projects; attached pictures go to the model as real images",
	}, "\n"),
	"v0.9.2": strings.Join([]string{
		"jin web: tools show / hide next to the usage line replaces the four folding tabs",
		"jin web: ask_user questions come one at a time with Next and Back",
		"jin web: wide code scrolls inside its block; no scroll bars, no band behind the input",
	}, "\n"),
	"v0.9.1": strings.Join([]string{
		"jin web: the Ctrl+O folding modes as tabs under the input; the choice is shared with the TUI",
		"jin web: pasted pictures show as thumbnails above the input",
		"jin web: a calmer sidebar, one line per session, search behind an icon",
	}, "\n"),
	"v0.9.0": strings.Join([]string{
		"jin web: the same jin in the browser, sharing projects, sessions and settings with the TUI",
		"Up to four panes, a command palette, /, #, @ and $ in the composer, pictures by paste or drop",
		"Run it with jin web; --port, --no-open and --cwd change where and how it starts",
	}, "\n"),
	"v0.8.4": strings.Join([]string{
		"Background tasks belong to jin itself: the task tool replaces jin async and its daemon",
		"All tool calls of one answer run at the same time; bash takes a dir",
		"tell_user: the agent writes to you mid-turn and suggests your next request in the input",
	}, "\n"),
	"v0.8.3": "The focused pane is blue and glows green only while its agent works; idle panes are darker",
	"v0.8.2": strings.Join([]string{
		"Each pane keeps its own todo list and ask_user question, also when it is not focused",
		"Pane borders glow around: green when focused, blue while working, purple for background tasks",
		"Esc goes back one panel; it closes only the first one",
	}, "\n"),
	"v0.8.1": strings.Join([]string{
		"/projects shows dots like /sessions: answering, unread answer or background tasks",
		"read tells the model it can open pictures, so screenshots get looked at",
		"Started in the home directory, jin no longer loads every global hook twice",
	}, "\n"),
	"v0.8.0": strings.Join([]string{
		"Projects: each session keeps its own directory; /projects switches, adds or removes one",
		"Up to four panes with one shared input and status; Tab or Alt+arrows moves the focus",
		"Global settings live in one /settings list",
		"One-line status bar; panels lost their tab bar; todos are the model's to write",
		"Voice, /copy, /diff, /links and /todo are gone; /tasks, /qa, /vertical and /horizontal are new names",
	}, "\n"),
	"v0.7.1": strings.Join([]string{
		"Voice input: /voice records the microphone and the words land in the draft",
		"/voice-provider: OpenAI-compatible endpoints or fal.run, checked before saving",
		"Enter in the / list runs the command at once; Tab inserts a token",
	}, "\n"),
	"v0.7.0": strings.Join([]string{
		"User messages get a padding line of background above and below",
		"README: new \"Why Jin?\" section; gallery gets skill, MCP and custom CLI examples",
		"Sharper async prompt: background work goes only through jin async",
	}, "\n"),
	"v0.6.9": strings.Join([]string{
		"Long sessions survive a full context: old tool output is pruned and compaction recovers",
		"Pasted text, images, /commands and #prompts are one token in the input; new /context and /diff",
		"jin -p gets --cwd, --provider, --max-cost, --max-turns and #prompts; jin sessions search",
		"Async tasks no longer hang on stdin; safer tools: atomic writes, bounded read, no stray terminal",
		"Cache-friendlier prompt order, an Anthropic breakpoint and prompt_cache_key; cache % in the status line",
	}, "\n"),
	"v0.6.3": "Input glow is longer and thicker, and stays visible as it wraps around",
	"v0.6.2": strings.Join([]string{
		"jin update installs the latest release; new sessions tell you when one is out",
		"Long text wraps between words: chat, input, todo and ask_user",
		"Custom themes in ~/.jin/themes (/theme: n new, e edit); panels have own colors",
		"/reset starts from scratch, /swap-config switches data folders; hook and prompt previews",
	}, "\n"),
	"v0.6.1": strings.Join([]string{
		"New provider kind: OpenAI Responses, with reasoning kept across turns",
		"Anthropic thinking signatures are replayed; prompt caching is on",
		"Silence timeout grows with reasoning effort: 90 s up to 10 min",
		"JIN_DEBUG=1 logs provider metadata: cache hits, timings, retries",
	}, "\n"),
	"v0.6": strings.Join([]string{
		"Ctrl+O has a new first mode: the last lines of bash output and edit diffs",
		"/undo restores the files of the agent's last turn; /rewind restarts from a message",
		"/theme (10 themes), /motion, /links; provider errors and stalled streams retry",
		"Project hooks in .jin/hooks, jin hooks add, jin export",
	}, "\n"),
}

// whatsNew is the intro section about this version, once per version. A
// fresh install has nothing to compare with, so it shows nothing.
func (a *app) whatsNew() (chatEntry, bool) {
	notes, ok := releaseNotes[a.version]
	if !ok || a.store == nil {
		return chatEntry{}, false
	}
	seen, err := a.store.Setting(keyReleaseSeen)
	if err != nil || seen == a.version {
		return chatEntry{}, false
	}
	_ = a.store.SetSetting(keyReleaseSeen, a.version)
	if seen == "" {
		if fresh, err := a.store.Empty(); err != nil || fresh {
			return chatEntry{}, false
		}
	}
	return section("What's new in "+a.version, notes), true
}
