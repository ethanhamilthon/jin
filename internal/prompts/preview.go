package prompts

import (
	"os"
	"strings"
)

// Preview is the start of a prompt's text on one line, for lists.
func Preview(name string) string {
	if body, ok := systemBody(name); ok {
		return Summary(body)
	}
	path, err := Path(name)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return Summary(string(data))
}

// Summary joins the first lines of text that carry words, without markdown
// heading marks, so a list can show what a file is about.
func Summary(text string) string {
	var words []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#>-*"))
		if line == "" || line == "---" {
			continue
		}
		words = append(words, strings.Fields(line)...)
		if len(words) > 40 {
			break
		}
	}
	return strings.Join(words, " ")
}
