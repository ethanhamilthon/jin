package headless

import (
	"errors"
	"fmt"
	"io"

	"jin/internal/session"
	"jin/internal/store"
)

// runProjectEdit adds, renames, archives or restores one project the way
// jin web does, then prints the project.
func runProjectEdit(action string, args []string, db *store.DB, dir string, out io.Writer) error {
	if action == "rename" && len(args) != 2 {
		return errors.New("usage: jin projects rename <path> <name>")
	}
	if action != "rename" && len(args) != 1 {
		return fmt.Errorf("usage: jin projects %s <path>", action)
	}
	path, err := session.ProjectPath(args[0], dir)
	if err != nil {
		return err
	}
	if action == "add" {
		if _, err := db.EnsureProject(path); err != nil {
			return err
		}
	} else if _, err := findProject(db, path); err != nil {
		return err
	}
	switch action {
	case "add", "restore":
		err = db.SetProjectArchived(path, false)
	case "archive":
		err = db.SetProjectArchived(path, true)
	case "rename":
		err = db.RenameProject(path, args[1])
	}
	if err != nil {
		return err
	}
	project, err := findProject(db, path)
	if err != nil {
		return err
	}
	printProjectRow(out, project)
	return nil
}

// findProject returns the registered project at path.
func findProject(db *store.DB, path string) (store.Project, error) {
	projects, err := db.Projects()
	if err != nil {
		return store.Project{}, err
	}
	for _, p := range projects {
		if p.Path == path {
			return p, nil
		}
	}
	return store.Project{}, fmt.Errorf("not a project, see jin projects list: %s", path)
}
