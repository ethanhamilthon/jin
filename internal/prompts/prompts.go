// Package prompts keeps reusable prompt files under ~/.jin/prompts. The file
// review/security.md is the prompt "review/security".
package prompts

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"jin/internal/paths"
)

const ext = ".md"

func root() (string, error) {
	return paths.Global("prompts")
}

// List returns every prompt name, sorted. A missing directory is an empty list.
func List() ([]string, error) {
	dir, err := root()
	if err != nil {
		return nil, err
	}
	var names []string
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ext) {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		if name := filepath.ToSlash(strings.TrimSuffix(rel, ext)); validName(name) {
			names = append(names, name)
		}
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	sort.Strings(names)
	return names, err
}

// Create makes an empty prompt file (and its folders) unless it exists.
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

// Delete removes a prompt and any folders it leaves empty.
func Delete(name string) error {
	path, err := Path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	top, err := root()
	if err != nil {
		return err
	}
	for dir := filepath.Dir(path); dir != top && os.Remove(dir) == nil; dir = filepath.Dir(dir) {
	}
	return nil
}
