//go:build desktop

package desktop

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeApp(t *testing.T, withCLI bool) (exe, cli string) {
	t.Helper()
	dir := t.TempDir()
	exe = filepath.Join(dir, "jin-desktop")
	cli = filepath.Join(dir, "jin")
	if err := os.WriteFile(exe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if withCLI {
		if err := os.WriteFile(cli, []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return exe, cli
}

func TestLinkCLICreatesAndRepeats(t *testing.T) {
	exe, cli := fakeApp(t, true)
	home := t.TempDir()
	for range 2 {
		if linked, err := linkCLI(exe, home); err != nil || !linked {
			t.Fatalf("linked=%v err=%v", linked, err)
		}
	}
	if got, _ := os.Readlink(filepath.Join(home, ".local", "bin", "jin")); got != cli {
		t.Fatalf("link points to %q", got)
	}
}

func TestLinkCLIReplacesFileAndForeignLink(t *testing.T) {
	exe, cli := fakeApp(t, true)
	home := t.TempDir()
	link := filepath.Join(linkDir(home), "jin")
	_ = os.MkdirAll(linkDir(home), 0o755)
	_ = os.WriteFile(link, []byte("old cli"), 0o755)
	if _, err := linkCLI(exe, home); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.Readlink(link); got != cli {
		t.Fatalf("file not replaced: %q", got)
	}
	_ = os.Remove(link)
	_ = os.Symlink("/elsewhere/jin", link)
	if _, err := linkCLI(exe, home); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.Readlink(link); got != cli {
		t.Fatalf("link not replaced: %q", got)
	}
}

func TestLinkCLISkipsWithoutBundledCLI(t *testing.T) {
	exe, _ := fakeApp(t, false)
	home := t.TempDir()
	if linked, err := linkCLI(exe, home); err != nil || linked {
		t.Fatalf("linked=%v err=%v", linked, err)
	}
	if _, err := os.Lstat(linkDir(home)); err == nil {
		t.Fatal("created a folder without a bundled CLI")
	}
}

func TestLinkCLIFailsOnUnwritableFolder(t *testing.T) {
	exe, _ := fakeApp(t, true)
	home := t.TempDir()
	_ = os.MkdirAll(linkDir(home), 0o755)
	_ = os.Chmod(linkDir(home), 0o500)
	defer os.Chmod(linkDir(home), 0o755)
	if os.Geteuid() == 0 {
		t.Skip("root can write anywhere")
	}
	if _, err := linkCLI(exe, home); err == nil {
		t.Fatal("expected an error")
	}
}

func TestInLoginPathAsksTheShell(t *testing.T) {
	shell := filepath.Join(t.TempDir(), "sh")
	_ = os.WriteFile(shell, []byte("#!/bin/sh\necho motd\necho /usr/bin:/home/u/.local/bin\n"), 0o755)
	t.Setenv("SHELL", shell)
	if in, known := inLoginPath("/home/u/.local/bin"); !in || !known {
		t.Errorf("in=%v known=%v", in, known)
	}
	if in, known := inLoginPath("/other"); in || !known {
		t.Errorf("in=%v known=%v", in, known)
	}
	t.Setenv("SHELL", "")
	if _, known := inLoginPath("/other"); known {
		t.Error("known without a shell")
	}
}

func get(n *notice, method, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/desktop/notice", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	n.ServeHTTP(w, r)
	return w
}

func TestNoticeRoute(t *testing.T) {
	n := newNotice()
	n.finish(nil)
	if w := get(n, http.MethodGet, ""); w.Code != http.StatusNoContent {
		t.Errorf("no problem: %d", w.Code)
	}
	if w := get(n, http.MethodPost, ""); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("post: %d", w.Code)
	}
	if w := get(n, http.MethodGet, "http://evil.example"); w.Code != http.StatusForbidden {
		t.Errorf("origin: %d", w.Code)
	}
	n = newNotice()
	n.finish(&noticeBody{Kind: "path", Dir: "~/.local/bin", Line: "export X"})
	w := get(n, http.MethodGet, shellOrigin)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"kind":"path"`) || !strings.Contains(w.Body.String(), "export X") {
		t.Errorf("json: %d %s", w.Code, w.Body.String())
	}
}

func TestNoticeWaitsForTheCheck(t *testing.T) {
	n := newNotice()
	go n.finish(&noticeBody{Kind: "link", Message: "m"})
	if w := get(n, http.MethodGet, ""); w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
}

func TestCheckCLIReportsAMissingPathEntry(t *testing.T) {
	exe, _ := fakeApp(t, true)
	home := t.TempDir()
	shell := filepath.Join(t.TempDir(), "sh")
	_ = os.WriteFile(shell, []byte("#!/bin/sh\necho /usr/bin\n"), 0o755)
	t.Setenv("SHELL", shell)
	if body := checkCLI(exe, home); body == nil || body.Kind != "path" {
		t.Fatalf("body = %+v", body)
	}
	_ = os.WriteFile(shell, []byte("#!/bin/sh\necho /usr/bin:"+linkDir(home)+"\n"), 0o755)
	if body := checkCLI(exe, home); body != nil {
		t.Fatalf("body = %+v", body)
	}
}
