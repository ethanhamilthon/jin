// Package datadir moves the whole jin data directory (~/.jin): /reset puts
// it aside so jin starts from scratch, /swap-config exchanges it with
// another one. Both run after the TUI closed the database.
package datadir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"jin/internal/paths"
)

// Current is the data directory of this build.
func Current() (string, error) { return paths.Global() }

// Expand turns "~/x" into an absolute path and cleans it.
func Expand(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("a folder is required")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return filepath.Abs(path)
}

// CheckReset tells whether the data directory can be moved to dest: dest
// must not exist yet and must lie outside the data directory.
func CheckReset(current, dest string) error {
	if inside(dest, current) {
		return errors.New("the folder must be outside " + current)
	}
	if _, err := os.Lstat(dest); err == nil {
		return errors.New(dest + " already exists; choose a new folder")
	}
	if _, err := os.Stat(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("parent folder: %w", err)
	}
	return nil
}

// Reset moves the data directory to dest. Jin creates a fresh one on start.
func Reset(current, dest string) error {
	if err := CheckReset(current, dest); err != nil {
		return err
	}
	return os.Rename(current, dest)
}

// CheckSwap tells whether other looks like a jin data directory that can
// take the place of current.
func CheckSwap(current, other string) error {
	if inside(other, current) || inside(current, other) {
		return errors.New("the folder must not contain or be inside " + current)
	}
	info, err := os.Stat(other)
	if err != nil || !info.IsDir() {
		return errors.New(other + " is not a folder")
	}
	if _, err := os.Stat(filepath.Join(other, "jin.db")); err != nil {
		return errors.New(other + " has no jin.db; it is not a jin data folder")
	}
	return nil
}

// Swap exchanges the two folders: other becomes the data directory and the
// current one takes the old place of other, so a second swap undoes it.
func Swap(current, other string) error {
	if err := CheckSwap(current, other); err != nil {
		return err
	}
	parking := current + ".swap-" + strconv.Itoa(os.Getpid())
	if err := rename(current, parking); err != nil {
		return err
	}
	if err := rename(other, current); err != nil {
		return errors.Join(err, rename(parking, current))
	}
	if err := rename(parking, other); err != nil {
		return errors.Join(err, rename(current, other), rename(parking, current))
	}
	return nil
}

var rename = os.Rename

func inside(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && (rel == "." || !strings.HasPrefix(rel, ".."))
}
