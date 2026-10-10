package ui

import (
	"path/filepath"

	"github.com/gdamore/tcell/v3"

	"jin/internal/store"
)

// addRow is the value of the "Add project" row, which no project has.
const addRow = ""

func (a *app) openProjects() { a.openProjectList(a.dir) }

// openProjectList is the first step of the picker: the projects that are not
// archived, in the store's order, then "Add project". current is the row the
// cursor starts on.
func (a *app) openProjectList(current string) {
	projects, err := a.store.Projects()
	if err != nil {
		a.report(err)
		return
	}
	options := make([]option, 0, len(projects)+1)
	for _, project := range projects {
		if !project.Archived {
			options = append(options, option{label: shortPath(project.Path), detail: projectNote(project), value: project.Path})
		}
	}
	options = append(options, option{label: "Add project", value: addRow})
	activity := a.loadProjectActivity()
	sel := a.openList("Projects", options, current, a.chooseProject)
	sel.hint = "Enter open · a add · r rename · x archive · / search"
	sel.dot = func(path string) (string, tcell.Style) { return a.projectMark(path, activity) }
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addProjectField() },
		'r': a.renameProject,
		'x': a.confirmArchiveProject,
	}
}

// chooseProject opens the sessions of a project, or the field that adds one.
func (a *app) chooseProject(path string) error {
	if path == addRow {
		a.addProjectField()
		return nil
	}
	a.openProjectSessions(path)
	return nil
}

// projectNote is the name of a renamed project. A name that is only the
// folder name is not repeated.
func projectNote(project store.Project) string {
	if project.Name == filepath.Base(project.Path) {
		return ""
	}
	return project.Name
}
