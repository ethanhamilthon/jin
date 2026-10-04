package ui

import (
	"strings"
	"unicode"
)

// slash is the autocomplete panel for a /command typed anywhere in the input.
type slash struct {
	sel   *selector
	start int
}

// closedToken remembers an autocomplete closed with Esc, so the token stays
// plain text until it changes. kind is the trigger: '/', '@' or '#'.
type closedToken struct {
	session string
	kind    byte
	start   int
	query   string
}

// forgetClosed drops the Esc memory once its token is gone from the input,
// so the same trigger typed again later opens its list.
func (a *app) forgetClosed(kind byte, start int, found bool) {
	d := a.closed
	if d.kind == kind && (!found || d.session != a.active.id || d.start != start) {
		a.closed = closedToken{}
	}
}

func isSlashNameCluster(cluster string) bool {
	runes := []rune(cluster)
	return len(runes) == 1 && (unicode.IsLetter(runes[0]) || unicode.IsDigit(runes[0]) || runes[0] == '-')
}

// slashAt finds the /token that ends at the cursor. The / must start the text
// or follow whitespace, so paths and "and/or" are never commands.
func slashAt(input []string, cursor int) (start int, query []string, ok bool) {
	for i := cursor - 1; i >= 0; i-- {
		switch cluster := input[i]; {
		case cluster == "/":
			if i == 0 || strings.TrimSpace(input[i-1]) == "" {
				return i, input[i+1 : cursor], true
			}
			return 0, nil, false
		case !isSlashNameCluster(cluster):
			return 0, nil, false
		}
	}
	return 0, nil, false
}

func (a *app) refreshSlash() {
	previous, previousStart := "", -1
	if a.slash != nil {
		previous, previousStart = a.slash.sel.current(), a.slash.start
	}
	a.slash = nil
	s := a.active
	if a.sel != nil || s.ask != nil || s.bash != nil || a.voice != nil {
		return
	}
	start, query, ok := slashAt(s.input, s.cursor)
	a.forgetClosed('/', start, ok)
	if !ok {
		return
	}
	typed := strings.ToLower(strings.Join(query, ""))
	if d := a.closed; d.session == s.id && d.kind == '/' && d.start == start && d.query == typed {
		return
	}
	var options []option
	for _, c := range slashCommands() {
		if strings.HasPrefix(c.name, typed) {
			options = append(options, option{label: "/" + c.name, detail: c.desc, value: c.name})
		}
	}
	if len(options) == 0 {
		return
	}
	icons := map[string]string{}
	for _, c := range slashCommands() {
		icons[c.name] = c.icon
	}
	sel := &selector{title: "Commands", options: options, mark: func(name string) string { return icons[name] }}
	if start == previousStart {
		sel.selectValue(previous)
	}
	a.slash = &slash{sel: sel, start: start}
}
