package store

import (
	"os"
)

func secureDBFiles(file string) error {
	for _, suffix := range []string{"", "-journal", "-wal", "-shm"} {
		path := file + suffix
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return &os.PathError{Op: "secure database", Path: path, Err: os.ErrInvalid}
		}
		if err := os.Chmod(path, 0o600); err != nil {
			return err
		}
	}
	return nil
}
