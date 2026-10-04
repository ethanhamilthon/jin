package ui

import (
	"os"
	"strings"

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
	if a.sel != nil || s.ask != nil || s.bashInput() || a.slash != nil {
		return
	}
	text := strings.Join(s.input, "")
	byteCursor := len(strings.Join(s.input[:s.cursor], ""))
	token, ok := files.Find(text, byteCursor)
	start, end := clusterOffset(s.input, token.Start), clusterOffset(s.input, token.End)
	a.forgetClosed('@', start, ok)
	if !ok {
		return
	}
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
