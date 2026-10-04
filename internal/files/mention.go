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
		closed := false
		if quoted {
			content++
			for j := content; j < len(text); j++ {
				if text[j] == '\\' && j+1 < len(text) {
					j++
					continue
				}
				if text[j] == '"' {
					end = j + 1
					closed = true
					break
				}
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
		if quoted && closed {
			rawEnd--
		}
		raw := text[content:rawEnd]
		if quoted {
			raw = unescape(raw)
		}
		return Token{Start: i, End: end, Raw: raw, Path: raw, Quoted: quoted}, true
	}
	return Token{}, false
}

func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
