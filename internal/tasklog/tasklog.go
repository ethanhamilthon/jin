// Package tasklog owns the files of background tasks: one log per task under
// ~/.jin/async, and for tasks that moved there from the bash tool, a file
// with the exit code. The bash tool, the async daemon and `jin async check`
// all use the same files, so a task can change hands without copying output.
package tasklog

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"jin/internal/paths"
)

// Dir is the folder of all task files.
func Dir() (string, error) { return paths.Global("async") }

// NewID makes a short random task id.
func NewID() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

// Files are the files of one task.
type Files struct {
	ID, Log, Exit string
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
		return Files{ID: id, Log: log, Exit: filepath.Join(dir, id+".exit")}, file.Close()
	}
	return Files{}, errors.New("could not reserve a task id")
}

// Within reports whether path is a file directly inside the task folder. The
// daemon trusts no other path from a client.
func Within(path string) bool {
	dir, err := Dir()
	if err != nil || path == "" {
		return false
	}
	rel, err := filepath.Rel(dir, filepath.Clean(path))
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..") && !strings.ContainsRune(rel, filepath.Separator)
}

// Remove deletes task files; a missing file is not an error.
func Remove(paths ...string) {
	for _, path := range paths {
		if path != "" {
			_ = os.Remove(path)
		}
	}
}
