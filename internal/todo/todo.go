// Package todo holds the session todo list: its items, the checks the model's
// updates must pass, and the text format used to edit the list in an editor.
package todo

import (
	"fmt"
	"regexp"
	"strings"
)

type Status string

const (
	Pending    Status = "pending"
	InProgress Status = "in_progress"
	Done       Status = "done"
)

type Item struct {
	Text   string `json:"text"`
	Status Status `json:"status"`
}

// Header is the comment written at the top of the editor file.
const Header = `<!-- One item per line: "- [ ] text" pending, "- [~] text" in progress, "- [x] text" done. Delete a line to remove the item. Save and close to apply. -->`

func (s Status) Valid() bool {
	return s == Pending || s == InProgress || s == Done
}

// Clean collapses line breaks in the text so an item always fits one line.
func Clean(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// Validate cleans the item texts and rejects empty texts and unknown
// statuses. An empty status means pending.
func Validate(items []Item) ([]Item, error) {
	out := make([]Item, 0, len(items))
	for i, item := range items {
		item.Text = Clean(item.Text)
		if item.Text == "" {
			return nil, fmt.Errorf("item %d: text is empty", i+1)
		}
		if item.Status == "" {
			item.Status = Pending
		}
		if !item.Status.Valid() {
			return nil, fmt.Errorf("item %d: unknown status %q (use pending, in_progress or done)", i+1, item.Status)
		}
		out = append(out, item)
	}
	return out, nil
}

// AllDone reports whether the list is not empty and every item is done.
func AllDone(items []Item) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item.Status != Done {
			return false
		}
	}
	return true
}

// Counts returns how many items are done and how many there are in total.
func Counts(items []Item) (done, total int) {
	for _, item := range items {
		if item.Status == Done {
			done++
		}
	}
	return done, len(items)
}

func Equal(a, b []Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func marker(s Status) string {
	switch s {
	case InProgress:
		return "[~]"
	case Done:
		return "[x]"
	}
	return "[ ]"
}

// Text renders the list for the model: one "- [ ] text" line per item.
func Text(items []Item) string {
	if len(items) == 0 {
		return "(empty)"
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = "- " + marker(item.Status) + " " + item.Text
	}
	return strings.Join(lines, "\n")
}

// Format renders the list as the editor file.
func Format(items []Item) string {
	var b strings.Builder
	b.WriteString(Header + "\n")
	for _, item := range items {
		b.WriteString("- " + marker(item.Status) + " " + item.Text + "\n")
	}
	return b.String()
}

var itemLine = regexp.MustCompile(`^\s*[-*]\s*\[( |~|x|X)\]\s+(\S.*)$`)

// Parse reads the editor file. Blank lines and single-line HTML comments are
// skipped; any other line that is not an item is an error.
func Parse(text string) ([]Item, error) {
	var items []Item
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || (strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->")) {
			continue
		}
		m := itemLine.FindStringSubmatch(line)
		if m == nil {
			return nil, fmt.Errorf(`line %d: expected "- [ ] text"`, i+1)
		}
		status := Pending
		switch m[1] {
		case "~":
			status = InProgress
		case "x", "X":
			status = Done
		}
		items = append(items, Item{Text: Clean(m[2]), Status: status})
	}
	return items, nil
}
