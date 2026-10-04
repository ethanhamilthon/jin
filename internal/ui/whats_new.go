package ui

import "strings"

const keyReleaseSeen = "release.seen"

// releaseNotes are the highlights of each version, shown once in the intro
// after an upgrade. Keep each to a few short lines.
var releaseNotes = map[string]string{
	"v0.8.1": "read tells the model it can open pictures, so screenshots get looked at",
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
