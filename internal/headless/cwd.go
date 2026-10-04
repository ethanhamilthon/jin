package headless

import (
	"fmt"
	"os"
	"path/filepath"
)

// enterDir makes dir the working directory of this process, so tools,
// commands and hooks run there, and returns its absolute path.
func enterDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return "", fmt.Errorf("--cwd %q is not a directory", dir)
	}
	if err := os.Chdir(abs); err != nil {
		return "", err
	}
	return abs, nil
}
