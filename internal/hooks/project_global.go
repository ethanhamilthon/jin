package hooks

import "os"

// isGlobalDir reports whether a project hook folder is the global one, as in
// a session started in the home directory, so its hooks are not loaded twice.
func isGlobalDir(projectDir string) bool {
	global, err := root()
	if err != nil {
		return false
	}
	a, err := os.Stat(projectDir)
	if err != nil {
		return false
	}
	b, err := os.Stat(global)
	return err == nil && os.SameFile(a, b)
}
