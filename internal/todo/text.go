package todo

import (
	"fmt"
	"regexp"
	"strings"
)

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
