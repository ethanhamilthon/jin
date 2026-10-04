//go:build unix

package async

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"jin/internal/store"
)

func TestCleanOldFilesKeepsRunningTasks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dir := t.TempDir()
	old := time.Now().Add(-2 * keepFiles)
	files := map[string]bool{"quiet.log": true, "done.log": false, "fresh.log": true, "done.exit": false}
	for name := range files {
		path := filepath.Join(dir, name)
		_ = os.WriteFile(path, []byte("x"), 0o600)
		if name != "fresh.log" {
			_ = os.Chtimes(path, old, old)
		}
	}
	err = db.AddAsyncTask(store.AsyncTask{ID: "quiet", SessionID: "s", Path: "/p", Command: "sleep",
		LogPath: filepath.Join(dir, "quiet.log"), StartedAt: old})
	if err != nil {
		t.Fatal(err)
	}
	cleanOldFiles(dir, db)
	for name, kept := range files {
		_, err := os.Stat(filepath.Join(dir, name))
		if (err == nil) != kept {
			t.Errorf("%s: kept = %v, want %v", name, err == nil, kept)
		}
	}
}

func TestOrdinaryTaskLogIsTrimmedWhileRunning(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.log")
	running := make(chan struct{})
	defer close(running)
	go trimWhileRunning(path, running)
	file, _ := os.Create(path)
	defer file.Close()
	if err := file.Truncate(80 << 20); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if info, _ := os.Stat(path); info.Size() < 80<<20 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Error("the log was not trimmed")
}
