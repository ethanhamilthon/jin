package ui

import (
	"context"
	"time"
)

// selector is the panel above the input. A list selector picks one option; a
// field selector edits free text in the input box. A list with actions binds
// them to plain letters and opens its search with "/"; a list without actions
// keeps its search open all the time.
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
	hint     string
	search   bool
	mark     func(value string) string
	actions  map[rune]func(value string)
	submit   func(value string) error
	onChoice func(value string, chosen int) error
	// onMove is told the value under the cursor whenever it changes;
	// onCancel runs when the list is closed with Esc.
	onMove   func(value string)
	onCancel func()
}

type option struct {
	label, detail, value string
	choices              []string
	chosen               int
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
	a.sel = sel
}

func (a *app) openList(title string, options []option, current string, submit func(string) error) *selector {
	sel := &selector{title: title, options: options, submit: submit}
	sel.selectValue(current)
	a.sel = sel
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
