package store

import "testing"

func TestSessionProjectsAndAllUnread(t *testing.T) {
	db := openTest(t)
	dir := t.TempDir()
	if err := db.TouchProvider("s1", dir, "m", "", "t", ""); err != nil {
		t.Fatal(err)
	}
	if err := db.SetUnread("s1", true); err != nil {
		t.Fatal(err)
	}
	projects, err := db.SessionProjects()
	want, _ := canonicalStoredPath(dir)
	if err != nil || projects["s1"] != want {
		t.Fatalf("projects = %v, %v; want s1 in %s", projects, err, want)
	}
	if unread, err := db.AllUnread(); err != nil || !unread["s1"] {
		t.Fatalf("unread = %v, %v", unread, err)
	}
}
