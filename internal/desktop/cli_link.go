//go:build desktop

package desktop

import (
	"fmt"
	"os"
	"path/filepath"

	"jin/internal/bundled"
)

func linkDir(home string) string { return filepath.Join(home, ".local", "bin") }

// linkCLI points ~/.local/bin/jin at the jin command inside the app. Whatever
// was at that path is replaced. A build without a bundled command (make
// desktop) does nothing, so it never touches the jin of the developer.
func linkCLI(exe, home string) (linked bool, err error) {
	target, ok := bundled.Path(exe)
	if !ok {
		return false, nil
	}
	dir := linkDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	link := filepath.Join(dir, "jin")
	if current, err := os.Readlink(link); err == nil && current == target {
		return true, nil
	}
	temp := fmt.Sprintf("%s.new-%d", link, os.Getpid())
	_ = os.Remove(temp)
	if err := os.Symlink(target, temp); err != nil {
		return false, err
	}
	if err := os.Rename(temp, link); err != nil {
		_ = os.Remove(temp)
		return false, err
	}
	return true, nil
}
