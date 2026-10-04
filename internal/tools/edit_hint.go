package tools

import (
	"regexp"
	"strconv"
	"strings"
)

var trailingSpace = regexp.MustCompile(`(?m)[ \t\r]+$`)

type normalization struct {
	cause string
	apply func(string) string
}

var normalizations = []normalization{
	{"line endings (CRLF versus LF)", func(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }},
	{"trailing whitespace", func(s string) string { return trailingSpace.ReplaceAllString(s, "") }},
	{"tabs versus spaces", func(s string) string { return strings.ReplaceAll(s, "\t", "    ") }},
}

// missHint explains a missed old_string when exactly one normalized
// comparison finds it. It never says how to apply the match: the model must
// copy the text from read.
func missHint(content, oldString string) string {
	for _, n := range normalizations {
		old := n.apply(oldString)
		if strings.TrimSpace(old) == "" {
			continue
		}
		text := n.apply(content)
		if strings.Count(text, old) != 1 {
			continue
		}
		line := 1 + strings.Count(text[:strings.Index(text, old)], "\n")
		return "; a match exists at line " + strconv.Itoa(line) + " when ignoring " + n.cause +
			", so the cause is " + n.cause +
			". Copy the text exactly from read, without the line-number prefix"
	}
	return ""
}
