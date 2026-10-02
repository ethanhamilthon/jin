package ui

import (
	"slices"
	"strings"

	"jin/internal/prompts"
)

// mention is the autocomplete panel for a #prompt typed in the input.
type mention struct {
	sel   *selector
	start int
}

// panel is whatever list is drawn above the input: a selector flow, or the
// prompt autocomplete.
func (a *app) panel() *selector {
	if a.sel != nil {
		return a.sel
	}
	if a.slash != nil {
		return a.slash.sel
	}
	if a.file != nil {
		return a.file.sel
	}
	if a.mention != nil {
		return a.mention.sel
	}
	return nil
}

// mentionAt finds the #token that ends at the cursor. The # must start the
// text or follow whitespace, matching how prompts are expanded on send.
func mentionAt(input []string, cursor int) (start int, query []string, ok bool) {
	for i := cursor - 1; i >= 0; i-- {
		switch cluster := input[i]; {
		case cluster == "#":
			if i == 0 || strings.TrimSpace(input[i-1]) == "" {
				return i, input[i+1 : cursor], true
			}
			return 0, nil, false
		case !isNameCluster(cluster):
			return 0, nil, false
		}
	}
	return 0, nil, false
}

func isNameCluster(cluster string) bool {
	runes := []rune(cluster)
	return len(runes) == 1 && prompts.IsNameRune(runes[0])
}

func (a *app) refreshMention() {
	previous, previousStart := "", -1
	if a.mention != nil {
		previous, previousStart = a.mention.sel.current(), a.mention.start
	}
	a.mention = nil
	s := a.active
	if a.sel != nil || s.ask != nil || s.bash != nil {
		return
	}
	start, query, ok := mentionAt(s.input, s.cursor)
	if !ok {
		return
	}
	if d := a.closed; d.session == s.id && d.kind == '#' && d.start == start && d.query == strings.Join(query, "") {
		return
	}
	// Only the prompts this session started with: their commands have run.
	names := make([]string, 0, len(s.promptBodies))
	for name := range s.promptBodies {
		names = append(names, name)
	}
	slices.Sort(names)
	options := make([]option, len(names))
	for i, name := range names {
		options[i] = option{label: name, value: name}
	}
	sel := &selector{title: "Prompts", options: options, query: slices.Clone(query)}
	if start == previousStart {
		sel.selectValue(previous)
	}
	visible := sel.visible()
	if len(visible) == 0 {
		return
	}
	if !sel.matches(sel.index) {
		sel.index = visible[0]
	}
	a.mention = &mention{sel: sel, start: start}
}
