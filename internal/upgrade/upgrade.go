// Package upgrade moves data left by an older jin to where this version
// expects it. Each step runs once and is recorded in the settings.
package upgrade

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"jin/internal/paths"
	"jin/internal/store"
)

const flag04 = "migrated.0_4"

// systemPrompts are the prompts jin 0.4 ships inside the binary. Jin 0.3
// copied them to ~/.jin/prompts as editable files.
var systemPrompts = []string{"plan", "review", "subagents"}

// Run applies every step that has not run yet.
func Run(db *store.DB) error {
	done, err := db.Flag(flag04)
	if err != nil || done {
		return err
	}
	if err := movePrompts04(); err != nil {
		return err
	}
	return db.SetFlag(flag04)
}

// movePrompts04 moves the old default prompt files to ~/.jin/prompts.bak, so
// the built-in prompts win without destroying edits the user made.
func movePrompts04() error {
	src, err := paths.Global("prompts")
	if err != nil {
		return err
	}
	dst, err := paths.Global("prompts.bak")
	if err != nil {
		return err
	}
	for _, name := range systemPrompts {
		from := filepath.Join(src, name+".md")
		if _, err := os.Stat(from); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		if err := os.MkdirAll(dst, 0o700); err != nil {
			return err
		}
		if err := os.Rename(from, uniquePath(filepath.Join(dst, name+".md"))); err != nil {
			return err
		}
	}
	return nil
}

// uniquePath keeps an earlier backup: name.md, then name.1.md, name.2.md.
func uniquePath(path string) string {
	ext := filepath.Ext(path)
	stem := path[:len(path)-len(ext)]
	candidate := path
	for i := 1; ; i++ {
		if _, err := os.Stat(candidate); errors.Is(err, fs.ErrNotExist) {
			return candidate
		}
		candidate = stem + "." + itoa(i) + ext
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}
