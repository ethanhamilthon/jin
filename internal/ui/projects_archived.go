package ui

// openArchivedProjects is the Archived projects row of /settings. Enter
// restores the project under the cursor.
func (a *app) openArchivedProjects() {
	projects, err := a.store.Projects()
	if err != nil {
		a.report(err)
		return
	}
	var options []option
	for _, project := range projects {
		if project.Archived {
			options = append(options, option{label: project.Name, detail: shortPath(project.Path), value: project.Path})
		}
	}
	sel := a.openList("Archived projects", options, "", func(path string) error {
		if err := a.store.SetProjectArchived(path, false); err != nil {
			return err
		}
		a.openArchivedProjects()
		return nil
	})
	sel.empty = "No archived projects"
	sel.hint = "Enter restore"
}
