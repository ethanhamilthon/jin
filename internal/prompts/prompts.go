// Package prompts keeps reusable prompt files under ~/.jin/prompts.
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

// ListInfo returns built-in and user prompts. System prompts come first.
func ListInfo() ([]Info, error) {
	result := append([]Info(nil), systemPromptList...)
	dir, err := root()
	if err != nil {
		return nil, err
	}
	var user []string
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ext) {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if name := filepath.ToSlash(strings.TrimSuffix(rel, ext)); validName(name) && !IsSystem(name) {
			user = append(user, name)
		}
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Strings(user)
	for _, name := range user {
		result = append(result, Info{Name: name, System: false})
	}
	return result, nil
}

func List() ([]string, error) {
	infos, err := ListInfo()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name
	}
	return names, nil
}

// Create makes an empty prompt file (and its folders) unless it exists.
func Create(name string) (string, error) {
	if IsSystem(name) {
		return "", ErrReserved
	}
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
	if IsSystem(name) {
		return ErrReserved
	}
	path, err := Path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	top, _ := root()
	for dir := filepath.Dir(path); dir != top && os.Remove(dir) == nil; dir = filepath.Dir(dir) {
	}
	return nil
}
