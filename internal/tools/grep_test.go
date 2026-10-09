package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func grepTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func runGrep(t *testing.T, dir, args string) string {
	t.Helper()
	out, err := Grep{dir: dir}.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("grep %s: %v", args, err)
	}
	return out
}

func TestGrepLines(t *testing.T) {
	dir := grepTree(t, map[string]string{"a.go": "package a\nfunc One() {}\n", "sub/b.go": "func Two() {}\nfunc One2() {}\n"})
	want := "a.go:2:func One() {}\nsub/b.go:2:func One2() {}"
	if got := runGrep(t, dir, `{"pattern":"func One"}`); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGrepPathGlobAndCase(t *testing.T) {
	dir := grepTree(t, map[string]string{"a.go": "Hello\n", "a.txt": "hello\n", "sub/b.go": "hello\n"})
	if got := runGrep(t, dir, `{"pattern":"hello","ignore_case":true,"glob":"*.go","mode":"files"}`); got != "a.go\nsub/b.go" {
		t.Errorf("glob by name: %q", got)
	}
	if got := runGrep(t, dir, `{"pattern":"hello","path":"sub"}`); got != "sub/b.go:1:hello" {
		t.Errorf("path: %q", got)
	}
	if got := runGrep(t, dir, `{"pattern":"hello","path":"a.txt"}`); got != "a.txt:1:hello" {
		t.Errorf("single file: %q", got)
	}
	if got := runGrep(t, dir, `{"pattern":"hello","glob":"sub/*.go","mode":"count"}`); got != "sub/b.go:1" {
		t.Errorf("glob by path, count: %q", got)
	}
}

func TestGrepContext(t *testing.T) {
	dir := grepTree(t, map[string]string{"f.txt": "1\nneedle\n3\n4\n5\n6\nneedle\n8\n"})
	want := "f.txt-1-1\nf.txt:2:needle\nf.txt-3-3\n--\nf.txt-6-6\nf.txt:7:needle\nf.txt-8-8"
	if got := runGrep(t, dir, `{"pattern":"needle","context":1}`); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGrepNoMatchSkipsAndErrors(t *testing.T) {
	dir := grepTree(t, map[string]string{"a.txt": "x\n", "bin.dat": "needle\x00\x01", ".hidden/h.txt": "needle\n", "node_modules/m.js": "needle\n"})
	got := runGrep(t, dir, `{"pattern":"needle"}`)
	if got != "No matches.\n[skipped 1 binary files]" {
		t.Errorf("got %q", got)
	}
	if _, err := (Grep{dir: dir}).Run(context.Background(), `{"pattern":"("}`); err == nil || !strings.Contains(err.Error(), "invalid pattern") {
		t.Errorf("bad pattern: %v", err)
	}
	for _, bad := range []string{`{}`, `{"pattern":"x","mode":"all"}`, `{"pattern":"x","context":11}`, `{"pattern":"x","glob":"["}`} {
		if _, err := (Grep{dir: dir}).Run(context.Background(), bad); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}

func TestGrepCutsOutput(t *testing.T) {
	dir := grepTree(t, map[string]string{"big.txt": strings.Repeat("needle "+strings.Repeat("x", 250)+"\n", 400)})
	got := runGrep(t, dir, `{"pattern":"needle"}`)
	if len(got) > grepMaxOutput+300 || !strings.Contains(got, "400 matches in 1 files in total") {
		t.Errorf("len %d, tail %q", len(got), got[max(0, len(got)-160):])
	}
	long := grepTree(t, map[string]string{"l.txt": "needle" + strings.Repeat("é", 400) + "\n"})
	if line := runGrep(t, long, `{"pattern":"needle"}`); !strings.HasSuffix(line, "…") || len(line) > 320 {
		t.Errorf("long line not cut: %d bytes", len(line))
	}
}

func TestGrepSkipsGitIgnored(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := grepTree(t, map[string]string{".gitignore": "ignored.txt\n", "ignored.txt": "needle\n", "kept.txt": "needle\n", "new.txt": "needle\n"})
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		t.Skip("git init failed")
	}
	if got := runGrep(t, dir, `{"pattern":"needle","mode":"files"}`); got != "kept.txt\nnew.txt" {
		t.Errorf("got %q", got)
	}
}
