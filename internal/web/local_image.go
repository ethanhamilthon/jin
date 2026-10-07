package web

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func resolveLocal(path, dir string) string {
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, rest)
	}
	if !filepath.IsAbs(path) && dir != "" {
		return filepath.Join(dir, path)
	}
	return filepath.Clean(path)
}

func readImageFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxImage {
		return nil, errors.New("not an image file")
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.HasPrefix(http.DetectContentType(data), "image/") {
		return nil, errors.New("not an image file")
	}
	return data, nil
}
