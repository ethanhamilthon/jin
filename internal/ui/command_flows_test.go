package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

// TestWorkViewHasNoSecondScreen checks that the entry keys of the removed
// Management view no longer open anything.
func TestWorkViewHasNoSecondScreen(t *testing.T) {
	a, _ := layoutApp(t)
	before := a.sel
	press(a, tcell.KeyF2)
	if a.sel != before {
		t.Fatalf("F2 must not open a screen of its own")
	}
	press(a, tcell.KeyLeft)
	if a.sel != before {
		t.Fatalf("Left with an empty draft must not open a screen of its own")
	}
}

// TestPanelCommandsOpenTheirOwnList checks that /sessions opens its own
// list, without a tab bar.
func TestPanelCommandsOpenTheirOwnList(t *testing.T) {
	a := startingApp(t)
	cmd, ok := slashByName("sessions")
	if !ok {
		t.Fatal("/sessions is missing")
	}
	cmd.run(a, "")
	if a.sel == nil || !strings.HasPrefix(a.sel.title, "Sessions · ") {
		t.Fatalf("/sessions = %+v, want the sessions list", a.sel)
	}
}

// TestSettingsListsEveryGlobalSetting checks that the global settings live
// behind one command, each row opening its own flow and nothing else.
func TestSettingsListsEveryGlobalSetting(t *testing.T) {
	a := startingApp(t)
	cmd, ok := slashByName("settings")
	if !ok {
		t.Fatal("/settings is missing")
	}
	cmd.run(a, "")
	if a.sel == nil || a.sel.title != "Settings" {
		t.Fatalf("/settings = %+v, want the settings list", a.sel)
	}
	var labels []string
	for _, opt := range a.sel.options {
		labels = append(labels, opt.label)
	}
	want := "Remote access,Sound,Swap config,Reset,Change editor,CLIProxyAPI,Tools,Scoped models,Motion,Prompts,Hooks,System prompt,Session titles,Archived projects"
	if got := strings.Join(labels, ","); got != want {
		t.Fatalf("settings rows = %q, want %q", got, want)
	}
	// Every row carries a description and a flow of its own.
	for _, entry := range a.settingsEntries() {
		if entry.detail == "" || entry.open == nil {
			t.Fatalf("the %s row is incomplete: %+v", entry.label, entry)
		}
	}
	// A row opens its flow: Prompts needs no provider and no background work.
	a.sel = nil
	a.openSettingsFlow()
	a.sel.selectValue("Prompts")
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Prompts" {
		t.Fatalf("the Prompts row = %+v, want the prompts list", a.sel)
	}
}

// TestProjectsCommandAddsAnExistingDirectory drives the a flow of the project
// list: an existing folder becomes a project, a missing one is refused.
func TestProjectsCommandAddsAnExistingDirectory(t *testing.T) {
	a := startingApp(t)
	dir := filepath.Join(a.dir, "..", filepath.Base(os.TempDir()), "jin-new-project")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	a.openProjects()
	a.sel.actions['a']("")
	if a.sel == nil || a.sel.title != "Project directory" {
		t.Fatalf("a must ask for a directory, got %+v", a.sel)
	}
	a.sel.query = clusters(dir)
	a.submitSelector()
	projects, err := a.store.Projects()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range projects {
		if sameProject(p.Path, dir) {
			return
		}
	}
	t.Fatalf("the directory was not registered: %+v", projects)
}

// TestPanelHeadersShowOnlyTheirOwnTitle checks that the tab row is gone: a
// panel header names just its own list.
func TestPanelHeadersShowOnlyTheirOwnTitle(t *testing.T) {
	a := startingApp(t)
	for _, tc := range []struct {
		open  func()
		title string
	}{
		{func() { a.openProjectSessions(a.dir) }, "Sessions"},
		{func() { a.openPromptsFlow() }, "Prompts"},
		{func() { a.showHooks("") }, "Hooks"},
	} {
		a.sel = nil
		tc.open()
		if a.sel == nil {
			t.Fatalf("%s did not open", tc.title)
		}
		a.drawRuleTitle(1, 60, a.sel)
		got := rowText(a.screen, 1, 60)
		if !strings.Contains(got, tc.title) {
			t.Errorf("header %q does not name %q", got, tc.title)
		}
		names := 0
		for _, name := range []string{"Sessions", "Prompts", "Hooks"} {
			names += strings.Count(got, name)
		}
		if names != 1 {
			t.Errorf("header %q carries %d list names, want only its own", got, names)
		}
	}
}

// TestSpaceTogglesPromptsAndHooks checks that Space runs the on/off action in
// both panels, the key the hint advertises.
func TestSpaceTogglesPromptsAndHooks(t *testing.T) {
	a := startingApp(t)
	a.openPromptsFlow()
	if a.sel.actions[' '] == nil {
		t.Fatal("prompts panel has no Space action")
	}
	toggled := 0
	a.sel.actions[' '] = func(string) { toggled++ }
	typeRune(a, " ")
	if toggled != 1 {
		t.Fatalf("Space ran the prompt toggle %d times, want 1", toggled)
	}
	a.sel = nil
	a.showHooks("")
	if a.sel.actions[' '] == nil {
		t.Fatal("hooks panel has no Space action")
	}
	toggled = 0
	a.sel.actions[' '] = func(string) { toggled++ }
	typeRune(a, " ")
	if toggled != 1 {
		t.Fatalf("Space ran the hook toggle %d times, want 1", toggled)
	}
}

// TestPanelHeadersCarryNoKeyHints checks that panel titles name the panel
// only: the keys live in the bottom line, not in the header.
func TestPanelHeadersCarryNoKeyHints(t *testing.T) {
	a := startingApp(t)
	panels := []struct {
		name string
		open func()
	}{
		{"theme", func() { a.openThemeFlow() }},
		{"sound", func() { a.openSoundFlow() }},
		{"tools", func() { a.openToolsFlow() }},
		{"prompts", func() { a.openPromptsFlow() }},
		{"hooks", func() { a.showHooks("") }},
		{"projects", func() { a.openProjects() }},
	}
	for _, panel := range panels {
		a.sel = nil
		panel.open()
		if a.sel == nil {
			t.Fatalf("%s panel did not open", panel.name)
		}
		title := a.sel.title
		if strings.ContainsAny(title, "←→↑↓") && strings.Contains(title, "·") {
			t.Errorf("%s title still lists keys: %q", panel.name, title)
		}
		a.drawRuleTitle(1, 60, a.sel)
		header := rowText(a.screen, 1, 60)
		for _, key := range []string{"Esc cancel", "Enter keep", "←/→ change"} {
			if strings.Contains(header, key) {
				t.Errorf("%s header carries the %q key hint: %q", panel.name, key, header)
			}
		}
	}
}

// TestBottomLineCarriesThePanelKeys checks that the keys a panel no longer
// puts in its title reach the input placeholder instead.
func TestBottomLineCarriesThePanelKeys(t *testing.T) {
	a, _ := layoutApp(t)
	for _, tc := range []struct{ hint, key string }{
		{"/ search · r reload · Enter keep", "Enter keep"},
		{"Enter open · / search", "Enter open"},
		{"Enter confirm · Esc cancel", "Enter confirm"},
		{"←/→ change · Enter hear · / search", "Enter hear"},
		{"←/→ on/off · applies to new sessions · / search", "on/off"},
	} {
		a.sel = &selector{title: "X", hint: tc.hint}
		box := a.inputBox()
		if !strings.Contains(box.placeholder, tc.key) {
			t.Errorf("placeholder %q does not carry %q", box.placeholder, tc.key)
		}
	}
}

// TestPromptAndHookPanelsKeepTheirActions checks that the prompt and hook
// panels, now behind /settings, still carry the add, delete, toggle and
// editor keys.
func TestPromptAndHookPanelsKeepTheirActions(t *testing.T) {
	a := startingApp(t)
	a.sel = nil
	a.openPromptsFlow()
	for _, key := range []rune{'a', 'd', 't', 'e'} {
		if a.sel.actions[key] == nil {
			t.Errorf("/prompts lost the %q action", key)
		}
	}
	a.sel = nil
	a.showHooks("")
	for _, key := range []rune{'a', 'p', 'd', 't', 'e'} {
		if a.sel.actions[key] == nil {
			t.Errorf("/hooks lost the %q action", key)
		}
	}
	if !strings.Contains(a.sel.hint, "project") {
		t.Errorf("hooks hint = %q, want the project key", a.sel.hint)
	}
}
