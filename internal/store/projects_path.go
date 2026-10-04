package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
)

func normalizedProjectPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" || !filepath.IsAbs(path) {
		return "", errors.New("project path must be absolute")
	}
	candidate := filepath.Clean(path)
	var suffix []string
	for {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			break
		}
		suffix = append(suffix, filepath.Base(candidate))
		candidate = parent
	}
	return filepath.Clean(path), nil
}

func canonicalStoredPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if !filepath.IsAbs(path) {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = absolute
	}
	return normalizedProjectPath(path)
}

func projectID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return "prj_" + hex.EncodeToString(raw[:]), nil
}
