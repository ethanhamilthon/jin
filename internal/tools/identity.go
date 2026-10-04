package tools

import (
	"path/filepath"
	"strings"
)

// fileIdentity resolves symlinks in an existing path prefix and preserves the
// missing suffix, giving new files the same identity through directory aliases.
func fileIdentity(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return absoluteIdentity(resolved)
	}
	candidate := strings.TrimRight(path, string(filepath.Separator))
	if candidate == "" && filepath.IsAbs(path) {
		candidate = string(filepath.Separator)
	}
	var suffix []string
	for {
		if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return absoluteIdentity(resolved)
		}
		parent, name := filepath.Split(candidate)
		if name == "" {
			if candidate == "." || candidate == string(filepath.Separator) {
				return path
			}
			candidate = strings.TrimRight(candidate, string(filepath.Separator))
			if candidate == "" {
				candidate = string(filepath.Separator)
			}
			continue
		}
		suffix = append(suffix, name)
		candidate = strings.TrimRight(parent, string(filepath.Separator))
		if candidate == "" {
			if filepath.IsAbs(parent) {
				candidate = string(filepath.Separator)
			} else {
				candidate = "."
			}
		}
	}
}

func absoluteIdentity(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}
