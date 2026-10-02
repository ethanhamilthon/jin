package ui

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v3"

	"jin/internal/files"
)

// maxFileCandidates caps the autocomplete list for one directory.
const maxFileCandidates = 200

// fileMention is the autocomplete panel for an @path typed in the input.
type fileMention struct {
	sel *selector
	// start and end are cluster indexes of the whole token: "@" through the
	// last cluster of the path (and the closing quote, if there is one).
	start, end int
	quoted     bool
	typed      string
	dirs       map[string]bool
}

// clusterOffset is the cluster index that starts at byte offset off.
func clusterOffset(input []string, off int) int {
	bytes := 0
	for i, c := range input {
		if bytes >= off {
			return i
		}
		bytes += len(c)
	}
	return len(input)
}

func (a *app) refreshFile() {
	previous, previousStart := "", -1
	if a.file != nil {
		previous, previousStart = a.file.sel.current(), a.file.start
	}
	a.file = nil
	s := a.active
	if a.sel != nil || s.ask != nil || s.bash != nil || a.slash != nil {
		return
	}
	text := strings.Join(s.input, "")
	byteCursor := len(strings.Join(s.input[:s.cursor], ""))
	token, ok := files.Find(text, byteCursor)
	if !ok {
		return
	}
	start, end := clusterOffset(s.input, token.Start), clusterOffset(s.input, token.End)
	if d := a.closed; d.session == s.id && d.kind == '@' && d.start == start && d.query == token.Raw {
		return
	}
	home, _ := os.UserHomeDir()
	candidates := files.Candidates(token.Raw, home, a.dir, maxFileCandidates)
	if len(candidates) == 0 {
		return
	}
	dirs := map[string]bool{}
	options := make([]option, len(candidates))
	for i, c := range candidates {
		value := files.Insert(c)
		dirs[value] = c.IsDir
		options[i] = option{label: files.Shorten(c.Path, max(10, a.width-12)), value: value}
	}
	sel := &selector{title: "Files", options: options, mark: func(value string) string {
		if dirs[value] {
			return "▸"
		}
		return " "
	}}
	if start == previousStart {
		sel.selectValue(previous)
	}
	a.file = &fileMention{sel: sel, start: start, end: end, quoted: token.Quoted, typed: token.Raw, dirs: dirs}
}

// fileKey lets the open list take navigation, completion and Esc. Enter
// completes unless the typed path is already a finished file, so a complete
// @file still sends.
func (a *app) fileKey(ev *tcell.EventKey) bool {
	f := a.file
	if f == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyUp:
		f.sel.move(-1)
	case tcell.KeyDown:
		f.sel.move(1)
	case tcell.KeyTab:
		a.acceptFile()
	case tcell.KeyEnter:
		current := f.sel.current()
		complete := !f.dirs[current] && (current == f.typed || current == `"`+f.typed+`"`)
		if a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0 || complete {
			return false
		}
		a.acceptFile()
	case tcell.KeyEscape:
		s := a.active
		a.closed = closedToken{session: s.id, kind: '@', start: f.start, query: f.typed}
		a.file = nil
	default:
		return false
	}
	return true
}

// acceptFile puts the highlighted path after the "@". A directory keeps the
// cursor inside the token so the list goes on into it; a file ends the token.
func (a *app) acceptFile() {
	s, f := a.active, a.file
	value := f.sel.current()
	if value == "" {
		return
	}
	completion := clusters(value)
	isDir := f.dirs[value]
	next := s.cursor
	if f.end > next {
		next = f.end
	}
	tail := s.input[next:]
	cursorAt := f.start + 1 + len(completion)
	if isDir && strings.HasSuffix(value, `"`) {
		cursorAt-- // stay before the closing quote
	}
	if !isDir && (len(tail) == 0 || strings.TrimSpace(tail[0]) != "") {
		completion = append(completion, " ")
		cursorAt = f.start + 1 + len(completion)
	}
	rest := append(completion, tail...)
	s.input = append(s.input[:f.start+1], rest...)
	s.cursor = cursorAt
}

// sendDraft sends the draft. Every @path that names a real file is replaced
// by its full path and listed in an <attached-files> block; the chat shows
// the draft as typed.
func (a *app) sendDraft(text string) {
	home, _ := os.UserHomeDir()
	clean, paths := files.Extract(text, home, a.dir)
	a.active.sendFiles(text, clean, files.Block(paths))
}
