package tasks

import (
	"os"
	"path/filepath"
	"time"

	"jin/internal/tasklog"
)

// keepLogs is how long the log of a task stays on disk after its last write.
const keepLogs = 7 * 24 * time.Hour

// CleanOldLogs removes task logs that nobody wrote to for a week. Logs of
// running tasks are skipped even when quiet.
func (m *Manager) CleanOldLogs() {
	dir, err := tasklog.Dir()
	if err != nil {
		return
	}
	inUse := map[string]bool{}
	for _, info := range m.Running("") {
		inUse[filepath.Clean(info.Log)] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil || entry.IsDir() || inUse[path] || time.Since(info.ModTime()) < keepLogs {
			continue
		}
		_ = os.Remove(path)
	}
}
