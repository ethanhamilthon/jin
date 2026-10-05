// Package tasklog owns the log files of commands: one log per command under
// ~/.jin/tasks. The bash tool writes to it, and a command that moves to the
// background keeps the same file, so no output is copied.
package tasklog

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"jin/internal/paths"
)

// Dir is the folder of all task files.
func Dir() (string, error) { return paths.Global("tasks") }

// NewID makes a short random task id.
func NewID() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

// Files are the files of one task.
type Files struct {
	ID, Log string
}

// New reserves an id and creates its empty log (mode 0600) in a folder that
// only the user can enter.
func New() (Files, error) {
	dir, err := Dir()
	if err != nil {
		return Files{}, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Files{}, err
	}
	for range 5 {
		id := NewID()
		log := filepath.Join(dir, id+".log")
		file, err := os.OpenFile(log, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return Files{}, err
		}
		return Files{ID: id, Log: log}, file.Close()
	}
	return Files{}, errors.New("could not reserve a task id")
}

// Remove deletes task files; a missing file is not an error.
func Remove(paths ...string) {
	for _, path := range paths {
		if path != "" {
			_ = os.Remove(path)
		}
	}
}
