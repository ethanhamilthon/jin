package ui

// settingEntry is one row of /settings: the name the list shows and the flow
// it opens.
type settingEntry struct {
	label  string
	detail string
	open   func()
}

// settingsEntries are the global settings, in the order /settings shows them.
func (a *app) settingsEntries() []settingEntry {
	return []settingEntry{
		{"Remote access", "Tailscale, pairing and connected devices", func() { a.openRemoteFlow() }},
		{"Sound", "Notification sound", func() { a.openSoundFlow() }},
		{"Swap config", "Use another jin data folder instead of the current one", func() { a.openSwapFlow() }},
		{"Reset", "Move all jin data aside and start from scratch", func() { a.openResetFlow() }},
		{"Change editor", "Choose the external editor", func() { a.chooseEditor(func() error { return nil }) }},
		{"CLIProxyAPI", "Install or update the subscription proxy", func() { a.openProxyFlow() }},
		{"Tools", "Switch agent tools on and off", func() { a.openToolsFlow() }},
		{"Scoped models", "Choose the models Ctrl+M cycles through", func() { a.openScopeFlow() }},
		{"Motion", "Animation speed of the glow, or off", func() { a.openMotionFlow() }},
		{"Prompts", "Reusable prompts: edit, add, delete", func() { a.openPromptsFlow() }},
		{"Hooks", "Hooks: edit, add, delete, on/off", func() { a.showHooks("") }},
		{"System prompt", "Edit the system, compact and handoff prompts", func() { a.openSystemPromptFlow() }},
		{"Session titles", "Named by a model after the first message", func() { a.openTitleFlow() }},
		{"Archived projects", "Restore a hidden project", func() { a.openArchivedProjects() }},
	}
}

// openSettingsFlow is /settings: one list of the global settings, each row
// opening its own flow.
func (a *app) openSettingsFlow() {
	entries := a.settingsEntries()
	options := make([]option, len(entries))
	for i, entry := range entries {
		options[i] = option{label: entry.label, detail: entry.detail, value: entry.label}
	}
	sel := a.openList("Settings", options, "", func(label string) error {
		for _, entry := range entries {
			if entry.label == label {
				entry.open()
				return nil
			}
		}
		return nil
	})
	sel.twoLines = true
	sel.hint = "Enter open · / search"
}
