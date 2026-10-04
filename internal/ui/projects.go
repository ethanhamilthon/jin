package ui

import "path/filepath"

func (a *app) openProjects() {
	projects, err := a.store.Projects()
	if err != nil {
		a.report(err)
		return
	}
	var options []option
	for _, project := range projects {
		options = append(options, option{label: filepath.Base(project.Path), detail: shortPath(project.Path), value: project.Path})
	}
	sel := a.openList("Projects", options, a.dir, a.switchProject)
	sel.twoLines = true
	sel.hint = "Enter switch · a add · d remove · / search"
	sel.empty = "No projects yet · a add"
	sel.mark = func(path string) string {
		if path == a.dir {
			return "●"
		}
		for _, s := range a.sessions {
			if s.path == path && (s.working || len(s.pending) > 0 || a.asyncRunning[s.id] > 0) {
				return "◐"
			}
		}
		return " "
	}
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
	name := filepath.Base(path)
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
