package ui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"
)

// layouts lists, per keyboard layout, its letters and the Latin keys they
// sit on. Only letters that differ from the layouts above need a row.
var layouts = [][2]string{
	// Russian ЙЦУКЕН; Kazakh and Belarusian share it.
	{"йцукенгшщзхъфывапролджэячсмитьбю", "qwertyuiop[]asdfghjkl;'zxcvbnm,."},
	// Ukrainian: і, ї, є, ґ replace ы, ъ, э and sit by the Enter key.
	{"іїєґ", "s]'\\"},
	// Belarusian: ў replaces щ.
	{"ў", "o"},
	// Kazakh letters sit on the digit row; і is taken by Ukrainian above.
	{"әңғүұқөһ", "245890-="},
	// Greek.
	{"ςερτυθιοπασδφγηξκλζχψωβνμάέήίόύώϊϋΐΰ", "wertyuiopasdfghjklzxcvbnmaehioyviyiy"},
}

// latinOf maps letters of non-Latin layouts to the Latin key they sit on,
// for terminals that report the typed letter but not the physical key.
var latinOf = func() map[rune]rune {
	m := map[rune]rune{}
	for _, layout := range layouts {
		latin := []rune(layout[1])
		for i, r := range []rune(layout[0]) {
			if _, taken := m[r]; !taken {
				m[r] = latin[i]
			}
		}
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
