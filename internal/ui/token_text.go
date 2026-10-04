package ui

import (
	"slices"
	"strings"
)

// renderTokens replaces every token marker in text by f of its token. Images
// are numbered in order first. An unknown marker is dropped.
func renderTokens(text string, f func(inputToken) string) string {
	var b strings.Builder
	images := 0
	for {
		open := strings.IndexRune(text, tokenOpen)
		if open < 0 {
			b.WriteString(text)
			return b.String()
		}
		end := strings.IndexRune(text[open:], tokenClose)
		if end < 0 {
			b.WriteString(text)
			return b.String()
		}
		end += open + len(string(tokenClose))
		b.WriteString(text[:open])
		if t, ok := tokenOf(text[open:end]); ok {
			if t.kind == tokenImage {
				images++
				t.label = imageLabel(images)
			}
			b.WriteString(f(t))
		}
		text = text[end:]
	}
}

func tokenLabel(t inputToken) string { return t.label }

// tokenPayload is the text a token stands for when the draft leaves the
// input as plain text: the editor, the clipboard, command arguments.
func tokenPayload(t inputToken) string {
	switch t.kind {
	case tokenPaste, tokenImage:
		return t.payload
	}
	return t.label
}

// tokenModelText is what the model reads in place of a token.
func tokenModelText(t inputToken) string {
	if t.kind == tokenImage {
		return strings.TrimSuffix(t.label, "]") + ": " + t.payload + "]"
	}
	return tokenPayload(t)
}

// promptNames lists the prompt tokens of text, once each, in order.
func promptNames(text string) []string {
	var names []string
	renderTokens(text, func(t inputToken) string {
		if t.kind == tokenPrompt && !slices.Contains(names, t.payload) {
			names = append(names, t.payload)
		}
		return ""
	})
	return names
}

func draftPayload(input []string) string {
	return renderTokens(strings.Join(input, ""), tokenPayload)
}

// renumberImages gives the image tokens of the input their numbers in order.
func renumberImages(input []string) {
	n := 0
	for _, cluster := range input {
		t, ok := tokenOf(cluster)
		if !ok || t.kind != tokenImage {
			continue
		}
		n++
		t.label = imageLabel(n)
		tokens.Lock()
		tokens.byMark[cluster] = t
		tokens.Unlock()
	}
}
