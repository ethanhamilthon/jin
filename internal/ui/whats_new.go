package ui

import "strings"

const keyReleaseSeen = "release.seen"

// releaseNotes are the highlights of each version, shown once in the intro
// after an upgrade. Keep each to a few short lines.
var releaseNotes = map[string]string{
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
