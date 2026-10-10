package ui

import "jin/internal/store"

// addProjectField asks for an existing folder and registers it as a project.
func (a *app) addProjectField() {
	baseDir := a.dir
	a.openDirField("Project directory", "", func(path string) error {
		dir, err := projectPath(path, baseDir)
		if err != nil {
			return err
		}
		if _, err := a.store.EnsureProject(dir); err != nil {
			return err
		}
		a.openProjectList(dir)
		return nil
	})
}

// renameProject asks for a new name, which starts as the current one.
func (a *app) renameProject(path string) {
	project, ok := a.projectAt(path)
	if !ok {
		return
	}
	a.openField("Project name", project.Name, false, func(name string) error {
		if err := a.store.RenameProject(path, name); err != nil {
			return err
		}
		a.openProjectList(path)
		return nil
	})
}

// confirmArchiveProject hides a project from the picker after a confirmation.
// Its sessions stay.
func (a *app) confirmArchiveProject(path string) {
	if _, ok := a.projectAt(path); !ok {
		return
	}
	options := []option{
		{label: "No", value: "no"},
		{label: "Yes, archive", detail: "Hide it from the picker. Sessions stay.", value: "yes"},
	}
	a.openList("Archive project "+shortPath(path)+"?", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := a.store.SetProjectArchived(path, true); err != nil {
				return err
			}
		}
		a.openProjectList(a.dir)
		return nil
	})
}

// projectAt finds a registered project by its path.
func (a *app) projectAt(path string) (store.Project, bool) {
	projects, err := a.store.Projects()
	if err != nil {
		a.report(err)
		return store.Project{}, false
	}
	for _, project := range projects {
		if project.Path == path {
			return project, true
		}
	}
	return store.Project{}, false
}
