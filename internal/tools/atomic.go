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

const maxLinkHops = 40

// resolveTarget follows a symlink chain one hop at a time to its final
// path, which may not exist yet, so a dangling link is filled in instead of
// replaced. The parent directory is resolved before each hop so a relative
// link target is read from the real directory.
func resolveTarget(path string) (string, error) {
	for range maxLinkHops {
		dir, base := filepath.Split(filepath.Clean(path))
		if dir == "" {
			dir = "."
		}
		realDir, err := filepath.EvalSymlinks(dir)
		if errors.Is(err, os.ErrNotExist) {
			return path, nil
		}
		if err != nil {
			return "", err
		}
		path = filepath.Join(realDir, base)
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) || (err == nil && info.Mode()&os.ModeSymlink == 0) {
			return path, nil
		}
		if err != nil {
			return "", err
		}
		link, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(link) {
			link = filepath.Join(realDir, link)
		}
		path = link
	}
	return "", errors.New("too many levels of symbolic links: " + path)
}
