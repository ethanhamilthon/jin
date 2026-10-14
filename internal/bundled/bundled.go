// Package bundled finds the jin command that the desktop app carries.
package bundled

import (
	"os"
	"path/filepath"
	"strings"
)

const desktopName = "jin-desktop"

// Path is the jin command next to the desktop binary exe. On macOS it is
// Contents/Resources/jin of the app (Contents/MacOS must not hold it: Jin and
// jin are one name on a case-insensitive disk); on Linux it is jin beside
// jin-desktop. ok is false when exe is not a desktop binary or the file is missing.
func Path(exe string) (path string, ok bool) {
	dir := filepath.Dir(exe)
	var candidate string
	switch {
	case filepath.Base(dir) == "MacOS" && filepath.Base(filepath.Dir(dir)) == "Contents":
		candidate = filepath.Join(filepath.Dir(dir), "Resources", "jin")
	case filepath.Base(exe) == desktopName:
		candidate = filepath.Join(dir, "jin")
	default:
		return "", false
	}
	info, err := os.Stat(candidate)
	return candidate, err == nil && info.Mode().IsRegular()
}

// Inside tells whether the jin command at path belongs to a desktop package.
// path must be the real path, with symbolic links resolved.
func Inside(path string) bool {
	if strings.Contains(filepath.ToSlash(path), ".app/Contents/Resources/") && filepath.Base(path) == "jin" {
		return true
	}
	if filepath.Base(path) != "jin" {
		return false
	}
	_, err := os.Stat(filepath.Join(filepath.Dir(path), desktopName))
	return err == nil
}

// AddToPath puts the folder of the bundled jin command first in PATH, so the
// commands of the agent find the jin of this version. It does nothing without
// a bundled command.
func AddToPath(exe string) {
	path, ok := Path(exe)
	if !ok {
		return
	}
	dir := filepath.Dir(path)
	if old := os.Getenv("PATH"); old != "" {
		dir += string(os.PathListSeparator) + old
	}
	os.Setenv("PATH", dir)
}
