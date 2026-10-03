package dyn

import (
	"context"
	"errors"
	"strings"
)

type piece struct {
	literal string
	command string
	isCmd   bool
}

// split cuts a text into literal pieces and commands. An empty {{}} and a {{
// that is never closed stay literal text.
func split(text string) []piece {
	var pieces []piece
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			pieces = append(pieces, piece{literal: lit.String()})
			lit.Reset()
		}
	}
	for i := 0; i < len(text); {
		switch {
		case strings.HasPrefix(text[i:], `\{{`):
			lit.WriteString("{{")
			i += 3
		case strings.HasPrefix(text[i:], "{{"):
			end := strings.Index(text[i+2:], "}}")
			if end < 0 {
				lit.WriteString(text[i:])
				i = len(text)
				break
			}
			command := strings.TrimSpace(text[i+2 : i+2+end])
			if command == "" {
				lit.WriteString(text[i : i+2+end+2])
			} else {
				flush()
				pieces = append(pieces, piece{command: command, isCmd: true})
			}
			i += 2 + end + 2
		default:
			lit.WriteByte(text[i])
			i++
		}
	}
	flush()
	return pieces
}

// Has reports whether a text holds a command, so callers can skip the work.
func Has(text string) bool {
	for _, p := range split(text) {
		if p.isCmd {
			return true
		}
	}
	return false
}

func marker(err error) string {
	if errors.Is(err, context.Canceled) {
		return "[command cancelled]"
	}
	return "[command failed: " + err.Error() + "]"
}
