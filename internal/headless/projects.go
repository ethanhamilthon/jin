package headless

import (
	"fmt"
	"slices"

	"jin/internal/store"
)

const projectsUsage = `usage:
  jin projects list [--all] [--format json]
  jin projects add <path>
  jin projects rename <path> <name>
  jin projects archive <path>
  jin projects restore <path>`

var projectEdits = []string{"add", "rename", "archive", "restore"}

// runProjects handles `jin projects ...`. Exit 0 on success, 1 on any error.
func runProjects(args []string, db *store.DB, dir string, io_ ioSet) int {
	var err error
	switch {
	case len(args) > 0 && args[0] == "list":
		err = runProjectsList(args[1:], db, io_.out)
	case len(args) > 0 && slices.Contains(projectEdits, args[0]):
		err = runProjectEdit(args[0], args[1:], db, dir, io_.out)
	default:
		fmt.Fprintln(io_.err, projectsUsage)
		return exitError
	}
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return exitError
	}
	return exitOK
}
