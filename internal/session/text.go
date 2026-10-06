package session

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// NewID is a random session id.
func NewID() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func FirstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

// Cut shortens text to limit runes, ending with "…" when it was longer.
func Cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}

func ShortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
