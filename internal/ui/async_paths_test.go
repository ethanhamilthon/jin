package ui

import (
	"testing"
	"time"

	"jin/internal/store"
)

func TestAsyncPollCoversEveryOpenProjectWithoutDuplicates(t *testing.T) {
	a := startingApp(t)
	other := t.TempDir()
	for _, entry := range []struct{ id, path string }{{"one", a.dir}, {"two", other}} {
		if err := a.store.Touch(entry.id, entry.path, "m", "", entry.id); err != nil {
			t.Fatal(err)
		}
		if err := a.store.AddAsyncEvent(entry.id, entry.path, entry.id+" result"); err != nil {
			t.Fatal(err)
		}
		if err := a.store.AddAsyncTask(store.AsyncTask{
			ID: entry.id, SessionID: entry.id, Path: entry.path,
			Command: "true", Status: "running", StartedAt: time.Now(),
		}); err != nil {
			t.Fatal(err)
		}
	}
	batch := a.pollAsyncPaths([]string{a.dir, other, a.dir})
	if len(batch.events) != 2 || batch.running["one"] != 1 || batch.running["two"] != 1 {
		t.Fatalf("wrong multi-project batch: %+v", batch)
	}
}

func TestWatchAsyncPathsKeepsHiddenSessionsAndReplacesOldSnapshots(t *testing.T) {
	a, _ := layoutApp(t)
	a.asyncPaths = make(chan []string, 1)
	a.sessions = map[string]*chatSession{
		"visible": {id: "visible", path: "/a"},
		"hidden":  {id: "hidden", path: "/b"},
		"same":    {id: "same", path: "/a"},
	}
	a.watchAsyncPaths()
	a.sessions["new"] = &chatSession{id: "new", path: "/c"}
	a.watchAsyncPaths()
	paths := <-a.asyncPaths
	if len(paths) != 3 || paths[0] != "/a" || paths[1] != "/b" || paths[2] != "/c" {
		t.Fatalf("wrong latest paths: %v", paths)
	}
}
