package files

import (
	"strings"
	"unicode"
)

type Token struct {
	Start, End int
	Raw        string
	Path       string
	Quoted     bool
}

func Find(text string, cursor int) (Token, bool) {
	if cursor < 0 || cursor > len(text) {
		return Token{}, false
	}
	for i := 0; i < len(text); i++ {
		if text[i] != '@' || (i > 0 && !unicode.IsSpace(rune(text[i-1]))) {
			continue
		}
		quoted := i+1 < len(text) && text[i+1] == '"'
		start := i + 1
		end := len(text)
		content := start
		if quoted {
			content++
			if j := strings.IndexByte(text[content:], '"'); j >= 0 {
				end = content + j + 1
			}
		} else {
			for j := start; j < len(text); j++ {
				if text[j] == ' ' || text[j] == '\t' || text[j] == '\n' || text[j] == '\r' {
					end = j
					break
				}
			}
		}
		if cursor < start || cursor > end {
			continue
		}
		rawEnd := end
		if quoted && end > content && text[end-1] == '"' {
			rawEnd--
		}
		raw := text[content:rawEnd]
		return Token{Start: i, End: end, Raw: raw, Path: raw, Quoted: quoted}, true
	}
	return Token{}, false
}
