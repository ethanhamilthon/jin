package ui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"
)

func handleInput(ev *tcell.EventKey, input *[]string, cursor *int) string {
	if !ev.Pressed() {
		return ""
	}
	switch ev.Key() {
	case tcell.KeyEnter:
		if ev.Modifiers()&tcell.ModShift != 0 {
			*input = append(*input, "")
			copy((*input)[*cursor+1:], (*input)[*cursor:])
			(*input)[*cursor] = "\n"
			*cursor++
			return ""
		}
		text := strings.TrimSpace(strings.Join(*input, ""))
		if text != "" {
			*input, *cursor = nil, 0
		}
		return text
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if *cursor > 0 {
			*input = append((*input)[:*cursor-1], (*input)[*cursor:]...)
			*cursor--
		}
	case tcell.KeyDelete:
		if *cursor < len(*input) {
			*input = append((*input)[:*cursor], (*input)[*cursor+1:]...)
		}
	case tcell.KeyLeft:
		if *cursor > 0 {
			*cursor--
		}
	case tcell.KeyRight:
		if *cursor < len(*input) {
			*cursor++
		}
	case tcell.KeyHome:
		*cursor = 0
	case tcell.KeyEnd:
		*cursor = len(*input)
	case tcell.KeyRune:
		if ev.Modifiers()&(tcell.ModCtrl|tcell.ModMeta) != 0 {
			break
		}
		text := strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, ev.Str())
		if text != "" {
			*input = append(*input, "")
			copy((*input)[*cursor+1:], (*input)[*cursor:])
			(*input)[*cursor] = text
			*cursor++
		}
	}
	return ""
}
