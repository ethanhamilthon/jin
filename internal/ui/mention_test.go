package ui

import (
	"strings"
	"testing"
)

func TestMentionAt(t *testing.T) {
	cases := []struct {
		text   string
		start  int
		query  string
		wantOK bool
	}{
		{"#rev", 0, "rev", true},
		{"hey #review/sec", 4, "review/sec", true},
		{"#", 0, "", true},
		{"a#rev", 0, "", false},
		{"# Title", 0, "", false},
		{"no mention", 0, "", false},
	}
	for _, c := range cases {
		input := clusters(c.text)
		start, query, ok := mentionAt(input, len(input))
		if ok != c.wantOK || (ok && (start != c.start || strings.Join(query, "") != c.query)) {
			t.Errorf("mentionAt(%q) = %d %q %v", c.text, start, strings.Join(query, ""), ok)
		}
	}
}

func TestAcceptMentionKeepsTextAroundIt(t *testing.T) {
	s := &chatSession{input: clusters("use #rev now"), cursor: len(clusters("use #rev"))}
	a := &app{active: s}
	sel := &selector{options: []option{{label: "review/security", value: "review/security"}}, query: clusters("rev")}
	a.mention = &mention{sel: sel, start: 4}
	a.acceptMention()
	if got := strings.Join(s.input, ""); got != "use #review/security now" {
		t.Errorf("input = %q", got)
	}
	if s.cursor != len(clusters("use #review/security")) {
		t.Errorf("cursor = %d", s.cursor)
	}
}
