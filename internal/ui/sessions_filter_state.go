package ui

import "strings"

func (sel *selector) filterText() string {
	if sel == nil {
		return ""
	}
	return strings.Join(sel.query, "")
}

func (sel *selector) applyFilter() {
	sel.applyFilterWithPrev(sel.current())
}

func (sel *selector) applyFilterWithPrev(prev string) {
	if sel.filter == nil {
		return
	}
	words := strings.Fields(strings.Join(sel.query, ""))
	if len(words) == 0 {
		sel.options = append([]option(nil), sel.allOptions...)
		sel.err = ""
		sel.restoreSelection(prev)
		return
	}
	opts, err := sel.filter(words)
	if err != nil {
		sel.err = err.Error()
		return
	}
	sel.err = ""
	sel.options = opts
	sel.restoreSelection(prev)
}

func (sel *selector) restoreSelection(prev string) {
	if len(sel.options) == 0 {
		sel.index = 0
		return
	}
	for i, opt := range sel.options {
		if opt.value == prev {
			sel.index = i
			return
		}
	}
	sel.index = 0
}

func (sel *selector) clearFilter() bool {
	if sel.filter == nil || len(sel.query) == 0 {
		return false
	}
	prev := sel.current()
	sel.query, sel.cursor = nil, 0
	sel.applyFilterWithPrev(prev)
	return true
}
