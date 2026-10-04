package hooks

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProjectDir is the folder of a project's own hooks, kept in its repository
// so a team shares them like AGENTS.md.
func ProjectDir(dir string) string { return filepath.Join(dir, ".jin", "hooks") }

// ProjectPath resolves a project hook name to its file.
func ProjectPath(dir, name string) (string, error) {
	name = strings.TrimSuffix(strings.TrimSpace(name), ext)
	if !validName(name) {
		return "", errBadName
	}
	return filepath.Join(ProjectDir(dir), name+ext), nil
}

// ListProject returns the project hook names of dir in alphabetical order.
func ListProject(dir string) ([]string, error) {
	if isGlobalDir(ProjectDir(dir)) {
		return nil, nil
	}
	entries, err := os.ReadDir(ProjectDir(dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	var names []string
	for _, entry := range entries {
		name, ok := strings.CutSuffix(entry.Name(), ext)
		if ok && !entry.IsDir() && validName(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, err
}

// CreateProject makes an empty project hook file unless it exists.
func CreateProject(dir, name string) (string, error) {
	path, err := ProjectPath(dir, name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	return path, file.Close()
}

// ProjectKey names a project hook in the list of disabled hooks. It is the
// file path, so the same name in two repositories is two hooks.
func ProjectKey(dir, name string) string {
	path, err := ProjectPath(dir, name)
	if err != nil {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}
