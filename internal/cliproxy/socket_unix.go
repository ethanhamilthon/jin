//go:build unix

package cliproxy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func socketPath(root string) string {
	if canonical, err := filepath.EvalSymlinks(root); err == nil {
		root = canonical
	}
	hash := sha256.Sum256([]byte(filepath.Clean(root)))
	return filepath.Join(fmt.Sprintf("/tmp/jin-proxy-%d", os.Getuid()), fmt.Sprintf("%x.sock", hash[:16]))
}

func socketDirectory(root string) error {
	dir := filepath.Dir(socketPath(root))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || info.Mode().Perm() != 0o700 || !ok || stat.Uid != uint32(os.Getuid()) {
		return fmt.Errorf("unsafe proxy control directory")
	}
	return nil
}
