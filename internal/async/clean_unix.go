//go:build unix

package async

import (
	"os"
	"path/filepath"
	"time"
)

// keepFiles is how long the log of a finished task stays on disk.
const keepFiles = 7 * 24 * time.Hour

// cleanOldFiles removes logs and exit files that nobody has touched for a
// week. A running task writes to its log, so it is never that old.
func cleanOldFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || entry.IsDir() || time.Since(info.ModTime()) < keepFiles {
			continue
		}
		switch filepath.Ext(entry.Name()) {
		case ".log", ".exit", ".tmp":
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}
