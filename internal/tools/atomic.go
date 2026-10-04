package tools

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// writeTarget names the file a write to path really changes. A symlink is
// followed to its existing target so the link survives; a broken link is
// refused. Any other path is returned untouched for the OS to resolve.
func writeTarget(path string) (string, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return path, nil
	}
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", errors.New("refusing to write through a broken symlink: " + path)
	}
	return target, nil
}

// parentOf cuts the last element by string, without Clean, so a ".." that
// follows a symlinked directory keeps its filesystem meaning.
func parentOf(path string) string {
	return path[:strings.LastIndexByte(path, os.PathSeparator)+1]
}

// writeFileAtomic replaces path with data through a temp file in the same directory, so a
// failure never leaves a half-written file. An existing file keeps its
// permission bits. It does not resolve symlinks: callers pass writeTarget's
// result.
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	temp, err := createSibling(path, mode)
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := fillTemp(temp, data, mode); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func createSibling(path string, mode os.FileMode) (*os.File, error) {
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		return nil, err
	}
	name := parentOf(path) + ".jin-" + hex.EncodeToString(suffix) + ".tmp"
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
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
