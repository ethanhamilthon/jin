package prompts

import (
	"os"
	"strings"
	"unicode/utf8"
)

const (
	openTag  = "<pasted-prompts>"
	closeTag = "</pasted-prompts>"
	intro    = "These are reusable prompts the user refers to by #name in the request below."
)

// Expand prepends the bodies of the prompts referenced as #name in text. Text
// without a known reference comes back unchanged.
func Expand(text string) string {
	names, _ := List()
	known := make(map[string]bool, len(names))
	for _, name := range names {
		known[name] = true
	}
	var block strings.Builder
	found := false
	for _, name := range references(text, known) {
		path, err := Path(name)
		if err != nil {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		found = true
		block.WriteString("<prompt name=\"" + name + "\">\n" + strings.TrimSpace(string(body)) + "\n</prompt>\n")
	}
	if !found {
		return text
	}
	return openTag + "\n" + intro + "\n" + block.String() + closeTag + "\n\n" + text
}

// Strip undoes Expand, returning what the user actually typed.
func Strip(content string) string {
	if !strings.HasPrefix(content, openTag) {
		return content
	}
	if _, rest, ok := strings.Cut(content, closeTag+"\n\n"); ok {
		return rest
	}
	return content
}

// references lists the known prompts mentioned as #name, once each, in order.
// A # only counts at the start of the text or after whitespace.
func references(text string, known map[string]bool) []string {
	var out []string
	seen := map[string]bool{}
	for i := 0; i < len(text); i++ {
		if text[i] != '#' || (i > 0 && !strings.ContainsRune(" \t\r\n", rune(text[i-1]))) {
			continue
		}
		token := nameToken(text[i+1:])
		if !known[token] {
			token = strings.TrimRight(token, ".-/_")
		}
		if known[token] && !seen[token] {
			seen[token] = true
			out = append(out, token)
		}
	}
	return out
}

func nameToken(s string) string {
	end := 0
	for end < len(s) {
		r, size := utf8.DecodeRuneInString(s[end:])
		if !IsNameRune(r) {
			break
		}
		end += size
	}
	return s[:end]
}
