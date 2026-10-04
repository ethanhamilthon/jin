package tools

import "path/filepath"

// fileIdentity is the one name of an existing file, whatever spelling
// reached it: symlinks and ".." keep their OS meaning. A path that does not
// exist is returned as written.
func fileIdentity(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	if abs, err := filepath.Abs(resolved); err == nil {
		return abs
	}
	return resolved
}
