package ui

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// pasteMark holds the place where a bracketed paste started until it ends.
const pasteMark = string(tokenOpen) + string(tokenClose)

func isLongPaste(text string) bool {
	return countLines(text) >= 3 || utf8.RuneCountInString(text) > 300
}

// replaceRange puts elements in place of input[start:end] and moves the
// cursor after them.
func replaceRange(input *[]string, cursor *int, start, end int, elements ...string) {
	*input = slices.Replace(*input, start, end, elements...)
	*cursor = start + len(elements)
}

// insertPaste puts pasted text at the cursor: long text as one token,
// short text as typed.
func insertPaste(input *[]string, cursor *int, text string) {
	if isLongPaste(text) {
		replaceRange(input, cursor, *cursor, *cursor, pasteToken(text))
		return
	}
	insertClusters(input, cursor, text)
}

func (a *app) pasteKey() {
	s := a.active
	text, image, ok := readClipboard()
	switch {
	case !ok:
	case image:
		replaceRange(&s.input, &s.cursor, s.cursor, s.cursor, imageToken(text))
	default:
		insertPaste(&s.input, &s.cursor, text)
	}
}

// beginPaste marks where a bracketed paste goes into the chat input.
func (a *app) beginPaste() {
	s := a.active
	if a.sel != nil || !s.ready || s.ask != nil || s.bash != nil || a.voice != nil {
		return
	}
	replaceRange(&s.input, &s.cursor, s.cursor, s.cursor, pasteMark)
}

// endPaste turns the text typed since beginPaste into a token when it is long.
func (a *app) endPaste() {
	s := a.active
	at := slices.Index(s.input, pasteMark)
	if at < 0 {
		return
	}
	end := max(at+1, min(s.cursor, len(s.input)))
	pasted := s.input[at+1 : end]
	text := strings.Join(pasted, "")
	if isLongPaste(text) {
		replaceRange(&s.input, &s.cursor, at, end, pasteToken(text))
		return
	}
	s.input = slices.Delete(s.input, at, at+1)
	if s.cursor > at {
		s.cursor--
	}
}

// tokenizeCommand turns a typed "/name" of a command into a token when a
// space follows it.
func (a *app) tokenizeCommand(ev *tcell.EventKey) {
	s := a.active
	if a.pasting || ev.Key() != tcell.KeyRune || ev.Str() != " " || ev.Modifiers()&(tcell.ModCtrl|tcell.ModMeta) != 0 {
		return
	}
	start, query, ok := slashAt(s.input, s.cursor)
	if !ok {
		return
	}
	name := strings.Join(query, "")
	if _, found := slashByName(name); found {
		replaceRange(&s.input, &s.cursor, start, s.cursor, commandToken(name))
	}
}
