package ui

import (
	"strconv"

	"jin/internal/store"
)

// openTitleFlow is the Session titles row of /settings: the same settings as
// the web Settings page, for the model that names sessions.
func (a *app) openTitleFlow() { a.showTitles("model") }

func (a *app) showTitles(current string) {
	t := a.cfg.Title
	options := []option{
		{label: "Model", detail: a.titleModelText(t), value: "model"},
		{label: "Effort", detail: effortOrDefault(t.Effort), value: "effort"},
		{label: "After", detail: afterText(t.After), value: "after"},
		{label: "Rename at message 4", detail: "names the session again once you have sent four messages", value: "refresh", choices: []string{"On", "Off"}, chosen: indexOf(!t.Refresh)},
		{label: "Prompt", detail: titlePromptText(t.Prompt), value: "prompt"},
	}
	sel := a.openList("Session titles", options, current, a.titleRow)
	sel.twoLines = true
	sel.keepOpen = true
	sel.hint = "←/→ on/off · Enter change · r session model · / search"
	sel.onChoice = a.changeRefreshTitle
	sel.actions = map[rune]func(string){
		'r': func(row string) {
			if row == "model" {
				a.resetTitleModel()
			}
		},
	}
}

// titleRow opens the change of a row; the switch row changes with ←/→.
func (a *app) titleRow(row string) error {
	switch row {
	case "model":
		a.pickTitleModel()
	case "effort":
		a.pickTitleEffort()
	case "after":
		a.openTitleAfter()
	case "prompt":
		return a.editTitlePrompt()
	}
	return nil
}

func (a *app) titleModelText(t store.TitleSettings) string {
	if t.Model == "" {
		return "The session's model"
	}
	name := t.Provider
	if entry, ok := a.configuredProvider(t.Provider); ok {
		name = entry.Name
	}
	return name + " · " + t.Model
}

func effortOrDefault(effort string) string {
	if effort == "" {
		return defaultEffort
	}
	return effort
}

func afterText(after int) string {
	if after == 0 {
		return "off"
	}
	return strconv.Itoa(after) + " · titled when you have sent this many messages"
}

func titlePromptText(prompt string) string {
	if prompt == store.DefaultTitlePrompt {
		return "Built-in prompt"
	}
	return oneLine(prompt)
}
