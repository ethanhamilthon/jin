package tools

import (
	"errors"
	"os"
	"path/filepath"
)

// writeFileAtomic replaces path with data through a temp file in the same
// directory, so a failure never leaves a half-written file. An existing
// file keeps its permission bits; a symlink is kept and its target replaced.
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	target, err := resolveTarget(path)
	if err != nil {
		return err
	}
	if info, err := os.Stat(target); err == nil {
		mode = info.Mode().Perm()
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".jin-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := fillTemp(temp, data, mode); err != nil {
		return err
	}
	return os.Rename(temp.Name(), target)
}

func fillTemp(temp *os.File, data []byte, mode os.FileMode) error {
	_, err := temp.Write(data)
	if err == nil {
		err = temp.Chmod(mode)
	}
	if err == nil {
		err = temp.Sync()
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	return err
}

func resolveTarget(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	return resolved, err
}
