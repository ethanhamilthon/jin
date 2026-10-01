package paths

import (
	"os"
	"path/filepath"
)

// mode is set by the production build; source builds use isolated dev data.
var mode = "dev"

func DirName() string {
	if mode == "prod" {
		return ".jin"
	}
	return ".jin-dev"
}

func Global(parts ...string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home, DirName()}, parts...)...), nil
}

func Local(parts ...string) string {
	return filepath.Join(append([]string{DirName()}, parts...)...)
}
