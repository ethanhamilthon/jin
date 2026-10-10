package ui

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"jin/internal/core"
	"jin/internal/export"
)

// exportSession writes the focused session as Markdown into its project
// folder and shows the file in the chat.
func (a *app) exportSession() {
	s := a.active
	if !s.persisted {
		a.report(errors.New("Nothing to export yet: send a message first"))
		return
	}
	rec, _, err := a.store.GetSession(s.id)
	if err != nil {
		a.report(err)
		return
	}
	messages, err := a.store.LoadMessages(s.id)
	if err != nil {
		a.report(err)
		return
	}
	text, err := export.Markdown(rec, messages)
	if err != nil {
		a.report(err)
		return
	}
	dir := s.path
	if dir == "" {
		dir = a.dir
	}
	path, err := writeExport(dir, rec.Title, s.id, text)
	if err != nil {
		a.report(err)
		return
	}
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Exported to " + path})
}

// writeExport creates a new file in dir named from the title and the start
// of the id. A name already taken gets a number, so nothing is overwritten.
func writeExport(dir, title, id string, text []byte) (string, error) {
	base := exportName(title) + "-" + shortID(id)
	for n := 1; n <= 1000; n++ {
		name := base + ".md"
		if n > 1 {
			name = base + "-" + strconv.Itoa(n) + ".md"
		}
		path := filepath.Join(dir, name)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, err = file.Write(text)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		return path, err
	}
	return "", errors.New("too many exports named " + base)
}

// exportName turns a title into a file name: no path separators or control
// characters, words joined by dashes, at most 60 characters.
func exportName(title string) string {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || strings.ContainsRune(`/\:*?"<>|`, r) {
			return ' '
		}
		return r
	}, title)
	name := []rune(strings.Join(strings.Fields(clean), "-"))
	if len(name) > 60 {
		name = name[:60]
	}
	if fitted := strings.Trim(string(name), ".-"); fitted != "" {
		return fitted
	}
	return "session"
}
