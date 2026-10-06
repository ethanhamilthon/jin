package session

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ProjectPath resolves a project directory typed by the user, relative to
// baseDir, to a readable directory without symlinks.
func ProjectPath(path, baseDir string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("Project path is required")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		if path == filepath.Join(home, "~") {
			path = home
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("Project path must be a directory")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Readdirnames(1)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return path, nil
}
