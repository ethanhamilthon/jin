package session

import "testing"

func TestAddAttachedListsPathsAndNames(t *testing.T) {
	paths, shown := addAttached([]string{"/a"}, "see this", []File{{Name: "r.pdf", Path: "/p/r.pdf"}})
	if len(paths) != 2 || paths[1] != "/p/r.pdf" || shown != "see this\n[file: r.pdf]" {
		t.Fatalf("paths %v, shown %q", paths, shown)
	}
	if _, shown := addAttached(nil, "plain ", nil); shown != "plain " {
		t.Fatalf("shown %q", shown)
	}
}
