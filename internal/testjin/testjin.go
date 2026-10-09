// Package testjin is for tests: the default system prompt runs `jin docs` and
// `jin hooks render`, and a test must not depend on which jin is installed.
package testjin

import (
	"os"
	"path/filepath"
	"testing"
)

const script = `#!/bin/sh
case "$1" in
docs) echo 'Jin documentation: pointer' ;;
hooks) [ "$2" = render ] && echo 'Hook text' ;;
esac
`

// Dir writes a fake jin into a new folder and returns the folder.
func Dir(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jin"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Env is the environment of jin with the fake jin first in PATH.
func Env(t testing.TB) []string {
	return append(os.Environ(), "PATH="+Dir(t)+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// OnPath puts the fake jin first in PATH for the rest of the test.
func OnPath(t testing.TB) {
	t.Helper()
	t.Setenv("PATH", Dir(t)+string(os.PathListSeparator)+os.Getenv("PATH"))
}
