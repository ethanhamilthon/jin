//go:build desktop

package desktop

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// inLoginPath tells whether dir is in the PATH of the login shell of the
// user. An app started from Finder has a short PATH of its own, so its
// environment says nothing. known is false when the shell gives no answer.
func inLoginPath(dir string) (in, known bool) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return false, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, shell, "-lc", "echo $PATH").Output()
	if err != nil {
		return false, false
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, entry := range filepath.SplitList(strings.TrimSpace(lines[len(lines)-1])) {
		if filepath.Clean(entry) == filepath.Clean(dir) {
			return true, true
		}
	}
	return false, true
}
