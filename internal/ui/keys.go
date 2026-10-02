package ui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"
)

// latinOf maps letters of non-Latin layouts to the Latin letter on the same
// key, for terminals that report the typed letter but not the physical key.
var latinOf = func() map[rune]rune {
	pairs := []string{
		"йцукенгшщзхъфывапролджэячсмитьбю", "qwertyuiop[]asdfghjkl;'zxcvbnm,.",
	}
	m := map[rune]rune{}
	latin := []rune(pairs[1])
	for i, r := range []rune(pairs[0]) {
		m[r] = latin[i]
	}
	return m
}()

// keyLetter is the lowercase Latin letter of a key whatever the keyboard
// layout, so shortcuts work the same with any language switched on.
func keyLetter(ev *tcell.EventKey) rune {
	if p := ev.Physical(); p >= tcell.KeyA && p <= tcell.KeyZ {
		return rune(p)
	}
	if ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ {
		return rune('a' + ev.Key() - tcell.KeyCtrlA)
	}
	runes := []rune(strings.ToLower(ev.Str()))
	if ev.Key() != tcell.KeyRune || len(runes) != 1 {
		return 0
	}
	if r, ok := latinOf[runes[0]]; ok {
		return r
	}
	return unicode.ToLower(runes[0])
}

// isCtrl matches Ctrl+letter in any layout. Meta counts too when cmd is set,
// so Cmd+C and Cmd+V work on macOS terminals that pass them through.
func isCtrl(ev *tcell.EventKey, letter rune, cmd bool) bool {
	mods := tcell.ModCtrl
	if cmd {
		mods |= tcell.ModMeta
	}
	legacy := ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ
	return (legacy || ev.Modifiers()&mods != 0) && keyLetter(ev) == letter
}
