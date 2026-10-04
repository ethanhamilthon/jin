package ui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// A token is one element of the input whose text is a private marker:
// tokenOpen, an id, tokenClose. Its label and payload live in the registry.
const (
	tokenOpen  = '\uE000'
	tokenClose = '\uE001'
)

type tokenKind uint8

const (
	tokenPaste tokenKind = iota
	tokenImage
	tokenCommand
	tokenPrompt
)

type inputToken struct {
	kind           tokenKind
	label, payload string
}

var tokens = struct {
	sync.Mutex
	next   int
	byMark map[string]inputToken
}{byMark: map[string]inputToken{}}

func newToken(t inputToken) string {
	tokens.Lock()
	defer tokens.Unlock()
	tokens.next++
	mark := string(tokenOpen) + strconv.Itoa(tokens.next) + string(tokenClose)
	tokens.byMark[mark] = t
	return mark
}

func tokenOf(cluster string) (inputToken, bool) {
	if !isTokenMark(cluster) {
		return inputToken{}, false
	}
	tokens.Lock()
	defer tokens.Unlock()
	t, ok := tokens.byMark[cluster]
	return t, ok
}

func isTokenMark(cluster string) bool {
	return strings.HasPrefix(cluster, string(tokenOpen))
}

func pasteToken(text string) string {
	n := countLines(text)
	label := fmt.Sprintf("[pasted text %d lines]", n)
	if n == 1 {
		label = "[pasted text 1 line]"
	}
	return newToken(inputToken{kind: tokenPaste, label: label, payload: text})
}

func imageLabel(n int) string { return fmt.Sprintf("[image %02d]", n) }

func imageToken(path string) string {
	return newToken(inputToken{kind: tokenImage, label: imageLabel(1), payload: path})
}

func commandToken(name string) string {
	return newToken(inputToken{kind: tokenCommand, label: "/" + name, payload: name})
}

func promptToken(name string) string {
	return newToken(inputToken{kind: tokenPrompt, label: "#" + name, payload: name})
}

func countLines(text string) int {
	return strings.Count(strings.TrimRight(text, "\n"), "\n") + 1
}
