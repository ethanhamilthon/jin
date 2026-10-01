package ui

import (
	"context"
	"strings"
	"time"
)

// selector is the panel above the input. A list selector picks one option and
// always keeps its search focused; a field selector edits free text in the
// input box. Actions are bound to Ctrl+letter so plain letters reach the search.
type selector struct {
	title    string
	options  []option
	index    int
	query    []string
	cursor   int
	field    bool
	secret   bool
	loading  bool
	err      string
	twoLines bool
	want     string
	empty    string
	tabbed   bool
	tab      int
	keepOpen bool
	mark     func(value string) string
	actions  map[rune]func(value string)
	submit   func(value string) error
}

type option struct {
	label, detail, value string
}

type loadResult struct {
	sel     *selector
	options []option
	err     error
}

func (a *app) openField(title, value string, secret bool, submit func(string) error) {
	sel := &selector{title: title, field: true, secret: secret, submit: submit}
	sel.query = clusters(value)
	sel.cursor = len(sel.query)
	a.sel, a.mode = sel, modeInsert
}

func (a *app) openList(title string, options []option, current string, submit func(string) error) *selector {
	sel := &selector{title: title, options: options, submit: submit}
	sel.selectValue(current)
	a.sel, a.mode = sel, modeInsert
	return sel
}

// openLoading opens an empty list and fills it from load in the background.
func (a *app) openLoading(title, current string, load func(context.Context) ([]option, error), submit func(string) error) *selector {
	sel := a.openList(title, nil, current, submit)
	sel.loading, sel.want = true, current
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()
		options, err := load(ctx)
		select {
		case a.loads <- loadResult{sel: sel, options: options, err: err}:
		case <-a.ctx.Done():
		}
	}()
	return sel
}

func (a *app) receiveLoad(result loadResult) {
	sel := result.sel
	if a.sel != sel {
		return
	}
	sel.loading = false
	if result.err != nil {
		sel.err = result.err.Error()
		return
	}
	sel.options = result.options
	sel.selectValue(sel.want)
}

func (sel *selector) selectValue(value string) {
	for i, opt := range sel.options {
		if opt.value == value {
			sel.index = i
			return
		}
	}
}

func (sel *selector) matches(i int) bool {
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
