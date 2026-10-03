package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeThemeFile saves t under a new name as a full theme file, so every
// color is there to edit. It never overwrites a file.
func writeThemeFile(name string, t theme) (string, error) {
	name = strings.TrimSpace(name)
	file := themeFileName(name)
	if file == "" {
		return "", errors.New("theme name needs a letter or a digit")
	}
	for _, existing := range allThemes() {
		if existing.name == name {
			return "", fmt.Errorf("theme %q already exists", name)
		}
	}
	dir, err := themesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	colors := map[string]string{}
	for key, slot := range themeSlots(&t) {
		if *slot != 0 {
			colors[key] = fmt.Sprintf("#%06X", *slot)
		}
	}
	data, err := json.MarshalIndent(themeFile{Name: name, Base: t.name, Colors: colors}, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, file+".json")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return path, err
}

// themeFileName is a file-safe form of a theme name.
func themeFileName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// customThemePath finds the file of a custom theme by its name.
func customThemePath(name string) (string, bool) {
	dir, err := themesDir()
	if err != nil {
		return "", false
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		t, err := parseTheme(data, strings.TrimSuffix(filepath.Base(path), ".json"))
		if err == nil && t.name == name {
			return path, true
		}
	}
	return "", false
}
