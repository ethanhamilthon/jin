package ui

import (
	"strings"

	"jin/internal/diff"
	"jin/internal/tools"
)

const diffBashNote = "Files changed through bash are not tracked."

// diffText is the full diff of every file the changes touched, one block per
// file from its first Before to its last After.
func diffText(changes []tools.Change) string {
	first := map[string]tools.Change{}
	last := map[string]string{}
	var order []string
	for _, c := range changes {
		if _, ok := first[c.Path]; !ok {
			first[c.Path] = c
			order = append(order, c.Path)
		}
		last[c.Path] = c.After
	}
	var b strings.Builder
	for _, path := range order {
		header := "--- " + path
		if !first[path].Existed {
			header = "--- " + path + " (new file)"
		}
		b.WriteString(header + "\n+++ " + path + "\n")
		for _, line := range diff.Lines(first[path].Before, last[path]) {
			b.WriteString(string(line.Op) + line.Text + "\n")
		}
		b.WriteString(finalNewlineNote(first[path].Before, last[path]))
		b.WriteString("\n")
	}
	return b.String() + diffBashNote + "\n"
}

// finalNewlineNote names a change of the last newline, which the line diff
// does not show.
func finalNewlineNote(before, after string) string {
	if before == "" || after == "" || strings.HasSuffix(before, "\n") == strings.HasSuffix(after, "\n") {
		return ""
	}
	if strings.HasSuffix(after, "\n") {
		return "\\ final newline added\n"
	}
	return "\\ final newline removed\n"
}
