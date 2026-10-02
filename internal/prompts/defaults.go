package prompts

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed defaults/*.md
var defaults embed.FS

// EnsureDefaults creates the default prompts that are missing. A file that
// exists is never touched, even when empty or edited, so a user's changes
// survive every update. A deleted default comes back at the next start.
func EnsureDefaults() error {
	dir, err := root()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	entries, err := defaults.ReadDir("defaults")
	if err != nil {
		return err
	}
	var failed error
	for _, entry := range entries {
		body, err := defaults.ReadFile("defaults/" + entry.Name())
		if err != nil {
			return err
		}
		if err := createExclusive(filepath.Join(dir, entry.Name()), body); err != nil && failed == nil {
			failed = err
		}
	}
	return failed
}

func createExclusive(path string, body []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, writeErr := file.Write(body)
	if closeErr := file.Close(); writeErr == nil {
		writeErr = closeErr
	}
	return writeErr
}
