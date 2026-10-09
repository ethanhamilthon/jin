package ui

import "github.com/gdamore/tcell/v3"

func (a *app) openProjects() {
	projects, err := a.store.Projects()
	if err != nil {
		a.report(err)
		return
	}
	var options []option
	for _, project := range projects {
		options = append(options, option{label: shortPath(project.Path), value: project.Path})
	}
	sel := a.openList("Projects", options, a.dir, a.switchProject)
	sel.hint = "Enter switch · a add · d remove · / search"
	sel.empty = "No projects yet · a add"
	activity := a.loadProjectActivity()
	sel.dot = func(path string) (string, tcell.Style) { return a.projectMark(path, activity) }
	sel.actions = map[rune]func(string){
		'a': func(string) {
			baseDir := a.dir
			a.openDirField("Project directory", "", func(path string) error {
				dir, err := projectPath(path, baseDir)
				if err != nil {
					return err
				}
				return a.switchProject(dir)
			})
		},
		'd': func(path string) { a.confirmRemoveProject(path) },
	}
}

// confirmRemoveProject unregisters a project after a confirmation. The open
// project is refused: its sessions and working directory are in use.
func (a *app) confirmRemoveProject(path string) {
	if path == "" {
		return
	}
	if sameProject(path, a.dir) {
		a.sel.err = "Cannot remove the open project"
		return
	}
	name := shortPath(path)
	options := []option{{label: "No", value: "no"}, {label: "Yes, remove", value: "yes"}}
	a.openList("Remove project "+name+"?", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := a.store.RemoveProject(path); err != nil {
				return err
			}
		}
		a.openProjects()
		return nil
	})
}
