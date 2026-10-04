package tools

import (
	"path/filepath"
	"strings"
)

func absoluteDir(dir string) string {
	if dir == "" {
		dir, _ = filepath.Abs(".")
		return dir
	}
	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}
	return dir
}

func toolPath(dir, path string) string {
	if dir == "" || filepath.IsAbs(path) {
		return path
	}
	if strings.HasSuffix(dir, string(filepath.Separator)) {
		return dir + path
	}
	return dir + string(filepath.Separator) + path
}
