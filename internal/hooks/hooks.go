// Package hooks keeps prompts added to the system prompt at session start.
package hooks

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"jin/internal/paths"
	"jin/internal/prompts"
)

const ext = ".md"

var errBadName = errors.New("name may use letters, digits, - _ . and cannot start with a dot")

func root() (string, error) {
	return paths.Global("hooks")
}

func validName(name string) bool {
	if name == "" || strings.HasPrefix(name, ".") {
		return false
	}
	for _, r := range name {
		if r == '/' || !prompts.IsNameRune(r) {
			return false
		}
	}
	return true
}

// Path resolves a hook name to its file.
func Path(name string) (string, error) {
	name = strings.TrimSuffix(strings.TrimSpace(name), ext)
	if !validName(name) {
		return "", errBadName
	}
	dir, err := root()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+ext), nil
}

// List returns every hook name in alphabetical order. A missing directory is
// an empty list.
func List() ([]string, error) {
	dir, err := root()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
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

// Create makes an empty hook file unless it exists.
func Create(name string) (string, error) {
	path, err := Path(name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	return path, file.Close()
}

func Delete(name string) error {
	path, err := Path(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
