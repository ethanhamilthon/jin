package store

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestEnsureProjectCanonicalizesAliasesAndMissingPaths(t *testing.T) {
	db := openTest(t)
	if _, err := db.EnsureProject("relative"); err == nil {
		t.Fatal("EnsureProject accepted a relative path")
	}
	real := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	first, err := db.EnsureProject(alias)
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.EnsureProject(real)
	if err != nil {
		t.Fatal(err)
	}
	expectedReal, _ := canonicalStoredPath(real)
	if first.ID != second.ID || first.Path != expectedReal {
		t.Fatalf("alias projects differ: %+v and %+v", first, second)
	}
	missing := filepath.Join(t.TempDir(), "not-created")
	expectedMissing, _ := canonicalStoredPath(missing)
	project, err := db.EnsureProject(missing)
	if err != nil {
		t.Fatal(err)
	}
	if project.Path != expectedMissing {
		t.Fatalf("missing path = %q, want %q", project.Path, expectedMissing)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("store created missing directory: %v", err)
	}
	projects, err := db.Projects()
	if err != nil || len(projects) != 2 {
		t.Fatalf("projects = %+v, %v", projects, err)
	}
}

func TestTouchAndRememberProjectEnforceSessionOwnership(t *testing.T) {
	db := openTest(t)
	first, second := t.TempDir(), t.TempDir()
	if err := db.TouchProvider("session", first, "m", "", "title", "p"); err != nil {
		t.Fatal(err)
	}
	project, err := db.EnsureProject(first)
	if err != nil {
		t.Fatal(err)
	}
	session, found, err := db.GetSession("session")
	if err != nil || !found || session.ProjectID != project.ID || session.Path != first {
		t.Fatalf("session = %+v, %v, %v", session, found, err)
	}
	if err := db.RememberProject(first, "session"); err != nil {
		t.Fatal(err)
	}
	updated, found, err := db.GetProject(project.ID)
	if err != nil || !found || updated.LastSession != "session" || updated.LastOpenedAt.IsZero() {
		t.Fatalf("project = %+v, %v, %v", updated, found, err)
	}
	if err := db.RememberProject(second, "session"); err == nil {
		t.Fatal("RememberProject bound a session to another project")
	}
	if err := db.Touch("session", second, "m2", "", "title"); err == nil {
		t.Fatal("Touch rebound a session to another project")
	}
	if err := db.RememberProject(first, "unknown"); err == nil {
		t.Fatal("RememberProject accepted an unknown session")
	}
	projects, err := db.Projects()
	if err != nil || len(projects) != 1 {
		t.Fatalf("failed cross-project calls changed registry: %+v, %v", projects, err)
	}
}

func TestConcurrentEnsureProjectIsIdempotent(t *testing.T) {
	first, second := openTwo(t)
	path := t.TempDir()
	var wg sync.WaitGroup
	ids := make(chan string, 40)
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		db := first
		if i%2 != 0 {
			db = second
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			project, err := db.EnsureProject(path)
			if err != nil {
				errs <- err
				return
			}
			ids <- project.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	var id string
	for got := range ids {
		if id == "" {
			id = got
		} else if got != id {
			t.Fatalf("concurrent EnsureProject returned IDs %q and %q", id, got)
		}
	}
	projects, err := first.Projects()
	if err != nil || len(projects) != 1 {
		t.Fatalf("projects = %+v, %v", projects, err)
	}
}

func TestRenameProjectKeepsThePath(t *testing.T) {
	db := openTest(t)
	dir := t.TempDir()
	if _, err := db.EnsureProject(dir); err != nil {
		t.Fatal(err)
	}
	if err := db.RenameProject(dir, "  Tracker  "); err != nil {
		t.Fatal(err)
	}
	projects, err := db.Projects()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range projects {
		if p.Path != canonicalOrSelf(dir) {
			continue
		}
		if p.Name != "Tracker" {
			t.Fatalf("name = %q, want Tracker", p.Name)
		}
		return
	}
	t.Fatal("project not found after renaming")
}

func canonicalOrSelf(path string) string {
	canonical, err := canonicalStoredPath(path)
	if err != nil {
		return path
	}
	return canonical
}

func TestArchivedProjectIsMarkedAndKept(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dir := t.TempDir()
	if _, err := db.EnsureProject(dir); err != nil {
		t.Fatal(err)
	}
	if err := db.SetProjectArchived(dir, true); err != nil {
		t.Fatal(err)
	}
	list, _ := db.Projects()
	if len(list) != 1 || !list[0].Archived {
		t.Fatalf("after archive: %+v", list)
	}
	if err := db.SetProjectArchived(dir, false); err != nil {
		t.Fatal(err)
	}
	if list, _ = db.Projects(); len(list) != 1 || list[0].Archived {
		t.Fatalf("after restore: %+v", list)
	}
}
