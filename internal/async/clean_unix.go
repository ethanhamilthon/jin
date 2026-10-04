//go:build unix

package async

import (
	"os"
	"path/filepath"
	"time"

	"jin/internal/store"
)

// keepFiles is how long the log of a finished task stays on disk.
const keepFiles = 7 * 24 * time.Hour

// cleanOldFiles removes logs and exit files that nobody has touched for a
// week. The files of running tasks stay, however quiet the task is.
func cleanOldFiles(dir string, db *store.DB) {
	tasks, err := db.RunningAsyncTasks("")
	if err != nil {
		return
	}
	inUse := map[string]bool{}
	for _, t := range tasks {
		inUse[filepath.Clean(t.LogPath)], inUse[filepath.Clean(t.ExitPath)] = true, true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		info, err := entry.Info()
		path := filepath.Join(dir, entry.Name())
		if err != nil || entry.IsDir() || inUse[path] || time.Since(info.ModTime()) < keepFiles {
			continue
		}
		switch filepath.Ext(entry.Name()) {
		case ".log", ".exit", ".tmp":
			_ = os.Remove(path)
		}
	}
}
