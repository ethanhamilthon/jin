package ui

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"jin/internal/store"
)

// registerProject makes a new folder a project and returns its stored path.
func registerProject(t *testing.T, a *app, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	project, err := a.store.EnsureProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	return project.Path
}

func storedProject(t *testing.T, a *app, path string) store.Project {
	t.Helper()
	projects, err := a.store.Projects()
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range projects {
		if project.Path == path {
			return project
		}
	}
	t.Fatalf("project %q is not stored", path)
	return store.Project{}
}

func optionOf(sel *selector, value string) (option, bool) {
	for _, opt := range sel.options {
		if opt.value == value {
			return opt, true
		}
	}
	return option{}, false
}

func TestProjectListOrdersUnarchivedProjectsAndEndsWithAdd(t *testing.T) {
	a := startingApp(t)
	one := registerProject(t, a, "one")
	two := registerProject(t, a, "two")
	three := registerProject(t, a, "three")
	if err := a.store.SetProjectArchived(two, true); err != nil {
		t.Fatal(err)
	}
	a.openProjects()
	var values []string
	for _, opt := range a.sel.options {
		values = append(values, opt.value)
	}
	want := []string{one, three}
	slices.Sort(want)
	want = append(want, addRow)
	if !slices.Equal(values, want) {
		t.Fatalf("rows = %q, want %q", values, want)
	}
	if last := a.sel.options[len(a.sel.options)-1].label; last != "Add project" {
		t.Fatalf("last row = %q, want Add project", last)
	}
}

func TestProjectKeysRenameAndArchive(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "app")
	a.openProjects()
	if a.sel.actions['d'] != nil {
		t.Fatal("d must not remove a project any more")
	}
	a.sel.actions['r'](dir)
	if a.sel == nil || a.sel.title != "Project name" || strings.Join(a.sel.query, "") != "app" {
		t.Fatalf("r must ask for the name, prefilled with it: %+v", a.sel)
	}
	a.sel.query = clusters("Renamed")
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Projects" {
		t.Fatalf("after the rename the list is back: %+v", a.sel)
	}
	if opt, _ := optionOf(a.sel, dir); opt.detail != "Renamed" {
		t.Fatalf("renamed row = %+v, want the name as detail", opt)
	}
	a.sel.actions['x'](dir)
	if a.sel == nil || !strings.HasPrefix(a.sel.title, "Archive project") {
		t.Fatalf("x must ask for a confirmation: %+v", a.sel)
	}
	a.sel.selectValue("yes")
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Projects" {
		t.Fatalf("after the archive the list is back: %+v", a.sel)
	}
	if _, ok := optionOf(a.sel, dir); ok {
		t.Fatal("an archived project is still listed")
	}
	if !storedProject(t, a, dir).Archived {
		t.Fatal("the project is not archived in the store")
	}
}

func TestAddRowOpensTheDirectoryField(t *testing.T) {
	a := startingApp(t)
	a.openProjects()
	a.sel.selectValue(addRow)
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Project directory" {
		t.Fatalf("the Add project row must ask for a directory: %+v", a.sel)
	}
}

func TestEnterOnProjectOpensItsSessionStep(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "app")
	a.openProjects()
	a.sel.selectValue(dir)
	a.submitSelector()
	if a.sel == nil || !strings.HasPrefix(a.sel.title, "Sessions · ") {
		t.Fatalf("Enter on a project must open its sessions: %+v", a.sel)
	}
	if a.sel.options[0].label != "New session" {
		t.Fatalf("first row = %q, want New session", a.sel.options[0].label)
	}
	if a.sel.back == nil || a.sel.back.title != "Projects" {
		t.Fatalf("the sessions step must return to the projects")
	}
}

func TestArchivedProjectsRestoreFromSettings(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "old")
	if err := a.store.SetProjectArchived(dir, true); err != nil {
		t.Fatal(err)
	}
	a.openSettingsFlow()
	a.sel.selectValue("Archived projects")
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Archived projects" || len(a.sel.options) != 1 {
		t.Fatalf("the settings row must list the archived project: %+v", a.sel)
	}
	a.submitSelector()
	if storedProject(t, a, dir).Archived {
		t.Fatal("Enter did not restore the project")
	}
	if a.sel == nil || a.sel.title != "Archived projects" || a.sel.empty != "No archived projects" {
		t.Fatalf("the list must stay open and empty: %+v", a.sel)
	}
}
