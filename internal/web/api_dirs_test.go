package web

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListDirsFoldersOnlyAndHiddenOnRequest(t *testing.T) {
	ts, root := newTestServer(t)
	for _, name := range []string{"b", "A", ".cache"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var list dirList
	if code := call(t, ts, "GET", "/api/dirs?path="+url.QueryEscape(root), "", &list); code != 200 {
		t.Fatalf("status %d", code)
	}
	if list.Path != root || list.Parent != filepath.Dir(root) || strings.Join(list.Dirs, ",") != "A,b" {
		t.Fatalf("list = %+v", list)
	}
	if code := call(t, ts, "GET", "/api/dirs?hidden=1&path="+url.QueryEscape(root), "", &list); code != 200 {
		t.Fatalf("status %d", code)
	}
	if strings.Join(list.Dirs, ",") != ".cache,A,b" {
		t.Fatalf("hidden list = %+v", list)
	}
	var failed struct{ Error string }
	if code := call(t, ts, "GET", "/api/dirs?path="+url.QueryEscape(filepath.Join(root, "file.txt")), "", &failed); code != 400 || failed.Error == "" {
		t.Fatalf("file path: status %d %+v", code, failed)
	}
}

func TestListDirsDefaultsToHomeAndStopsAtRoot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	list, err := listDirs("", false)
	if err != nil {
		t.Fatal(err)
	}
	if resolved, _ := filepath.EvalSymlinks(home); list.Path != resolved {
		t.Fatalf("home = %s, want %s", list.Path, resolved)
	}
	root, err := listDirs("/", false)
	if err != nil {
		t.Fatal(err)
	}
	if root.Path != "/" || root.Parent != "" {
		t.Fatalf("root = %+v", root)
	}
}
