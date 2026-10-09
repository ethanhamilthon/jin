package web

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// inProject resolves rel inside the project dir. A path that leaves the
// project, by "..", an absolute path or a symlink, is refused.
func inProject(dir, rel string) (string, error) {
	if dir == "" || !filepath.IsAbs(dir) {
		return "", errors.New("the project folder is missing")
	}
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(rel) {
		return "", errors.New("the path must be relative to the project")
	}
	full, err := filepath.EvalSymlinks(filepath.Join(root, filepath.Clean("/"+rel)))
	if err != nil {
		return "", err
	}
	if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		return "", errors.New("the path is outside the project")
	}
	return full, nil
}
