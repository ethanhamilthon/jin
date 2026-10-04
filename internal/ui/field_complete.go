package ui

import (
	"os"
	"strings"

	"jin/internal/files"
)

// completeDir lists the directories a typed path can continue into. Each
// candidate carries the whole text to put back in the field, so accepting one
// keeps the typed prefix and adds a level.
func (a *app) completeDir(query string) []option {
	home, _ := os.UserHomeDir()
	candidates := files.Candidates(query, home, a.dir, maxFileCandidates)
	options := make([]option, 0, len(candidates))
	for _, c := range candidates {
		if !c.IsDir {
			continue
		}
		options = append(options, option{label: c.Name + "/", value: c.Path})
	}
	return options
}

// openDirField opens a field that completes directory paths as they are typed.
func (a *app) openDirField(title, value string, submit func(string) error) {
	a.openField(title, value, false, submit)
	sel := a.sel
	sel.complete = a.completeDir
	a.refreshFieldCandidates(sel)
}

func (a *app) refreshFieldCandidates(sel *selector) {
	if sel.complete == nil {
		return
	}
	previous := sel.candidate()
	sel.cands = sel.complete(strings.Join(sel.query, ""))
	sel.indexCandidate(previous)
}

// candidate is the completion under the highlight, or "" when there is none.
func (sel *selector) candidate() string {
	if sel.candIdx < 0 || sel.candIdx >= len(sel.cands) {
		return ""
	}
	return sel.cands[sel.candIdx].value
}

// indexCandidate keeps the highlight on the same directory while it is still
// offered, and on the first one otherwise.
func (sel *selector) indexCandidate(value string) {
	sel.candIdx = 0
	if value == "" {
		return
	}
	for i, opt := range sel.cands {
		if opt.value == value {
			sel.candIdx = i
			return
		}
	}
}

// moveCandidate moves the completion highlight without leaving the field.
func (sel *selector) moveCandidate(delta int) {
	if len(sel.cands) == 0 {
		return
	}
	sel.candIdx = (sel.candIdx + delta + len(sel.cands)) % len(sel.cands)
}

// acceptCandidate puts the highlighted directory in the field, so the next
// keystrokes complete one level deeper.
func (a *app) acceptCandidate(sel *selector) {
	value := sel.candidate()
	if value == "" {
		return
	}
	sel.query, sel.cursor = clusters(value), len(clusters(value))
	a.refreshFieldCandidates(sel)
}
