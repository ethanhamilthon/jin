package ui

import "strings"

func (sel *selector) selectValue(value string) {
	for i, opt := range sel.options {
		if opt.value == value {
			sel.index = i
			return
		}
	}
}

func (sel *selector) matches(i int) bool {
	if i < 0 || i >= len(sel.options) || sel.options[i].header {
		return false
	}
	query := strings.ToLower(strings.TrimSpace(strings.Join(sel.query, "")))
	return strings.Contains(strings.ToLower(sel.options[i].label), query)
}

func (sel *selector) visible() []int {
	var out []int
	for i := range sel.options {
		if sel.matches(i) {
			out = append(out, i)
		}
	}
	return out
}

// rows lists what the panel draws: the matching options, and each group
// header in front of its first matching row.
func (sel *selector) rows() []int {
	var out []int
	header := -1
	for i, opt := range sel.options {
		switch {
		case opt.header:
			header = i
		case sel.matches(i):
			if header >= 0 {
				out = append(out, header)
				header = -1
			}
			out = append(out, i)
		}
	}
	return out
}

// current is the value under the cursor, or "" when nothing is selectable.
func (sel *selector) current() string {
	if sel.loading || len(sel.options) == 0 || !sel.matches(sel.index) {
		return ""
	}
	return sel.options[sel.index].value
}

// move steps through the visible options and wraps around at both ends, so
// Up from the first option lands on the last.
func (sel *selector) move(direction int) {
	visible := sel.visible()
	if len(visible) == 0 {
		return
	}
	position := -1
	for i, idx := range visible {
		if idx == sel.index {
			position = i
		}
	}
	switch {
	case position < 0 && direction < 0:
		position = len(visible) - 1
	case position < 0:
		position = 0
	default:
		position = (position + direction + len(visible)) % len(visible)
	}
	sel.index = visible[position]
}

func (sel *selector) hasActions() bool { return len(sel.actions) > 0 }

// searching tells whether typed text goes to the search box.
func (sel *selector) searching() bool {
	return sel.field || !sel.hasActions() || sel.search
}

func (sel *selector) closeSearch() {
	sel.search, sel.query, sel.cursor = false, nil, 0
}
