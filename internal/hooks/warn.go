package hooks

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// readBody returns the trimmed text of a hook file. A file that does not
// exist reads as empty; any other error comes back as a warning.
func readBody(path string) (body, warning string) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", ""
	}
	if err != nil {
		return "", "cannot read hook: " + err.Error()
	}
	return strings.TrimSpace(string(data)), ""
}

func listWarning(err error) []string {
	if err == nil {
		return nil
	}
	return []string{"cannot list hooks: " + err.Error()}
}
