package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"jin/internal/paths"
)

// themeFile is ~/.jin/themes/<file>.json. Every color is optional: a missing
// one comes from base, a built-in theme (Jin Original when empty).
type themeFile struct {
	Name   string            `json:"name"`
	Base   string            `json:"base,omitempty"`
	Colors map[string]string `json:"colors"`
}

// themeSlots maps the keys of a theme file to the fields of a theme.
func themeSlots(t *theme) map[string]*uint32 {
	return map[string]*uint32{
		"bg": &t.bg, "fg": &t.fg, "text": &t.text, "muted": &t.muted, "argument": &t.argument,
		"detail": &t.detail, "dim": &t.dim, "border": &t.border, "raised": &t.raised,
		"status": &t.status, "on_status": &t.onStatus, "status_title": &t.statusTitle,
		"accent": &t.accent, "green": &t.green, "amber": &t.amber, "red": &t.red,
		"purple": &t.purple, "pink": &t.pink, "teal": &t.teal,
		"panel": &t.panel, "todo_panel": &t.todo, "ask_panel": &t.ask,
		"slash_panel": &t.slash, "files_panel": &t.files, "mention_panel": &t.mention,
	}
}

func themesDir() (string, error) { return paths.Global("themes") }

// parseTheme reads one theme file. Unknown keys and bad colors are errors,
// so a typo does not pass silently.
func parseTheme(data []byte, fallbackName string) (theme, error) {
	var file themeFile
	if err := json.Unmarshal(data, &file); err != nil {
		return theme{}, err
	}
	t := builtinTheme(file.Base)
	t.name = strings.TrimSpace(file.Name)
	if t.name == "" {
		t.name = fallbackName
	}
	slots := themeSlots(&t)
	for key, value := range file.Colors {
		slot, ok := slots[key]
		if !ok {
			return theme{}, fmt.Errorf("unknown color %q", key)
		}
		hex := strings.TrimPrefix(strings.TrimSpace(value), "#")
		n, err := strconv.ParseUint(hex, 16, 32)
		if err != nil || len(hex) != 6 {
			return theme{}, fmt.Errorf("color %s: %q is not #RRGGBB", key, value)
		}
		// 0 means "derive" for the panel slots, so pure black is nudged.
		*slot = max(uint32(n), 1)
	}
	return t, nil
}

// customThemes loads every theme file, sorted by name. A broken file is
// reported in errs and skipped.
func customThemes() (list []theme, errs []string) {
	dir, err := themesDir()
	if err != nil {
		return nil, nil
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	sort.Strings(files)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err == nil {
			var t theme
			if t, err = parseTheme(data, strings.TrimSuffix(filepath.Base(path), ".json")); err == nil {
				list = append(list, t)
				continue
			}
		}
		errs = append(errs, filepath.Base(path)+": "+err.Error())
	}
	return list, errs
}
