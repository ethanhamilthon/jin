package ui

import (
	"strings"
	"testing"
)

func TestTruncateLeftKeepsTheEnd(t *testing.T) {
	cases := []struct {
		text  string
		width int
		want  string
	}{
		{"~/projects/jin", 20, "~/projects/jin"},
		{"~/projects/jin", 8, "…cts/jin"},
		{"~/projects/jin", 1, "…"},
		{"~/projects/jin", 0, ""},
	}
	for _, c := range cases {
		if got := truncateLeft(c.text, c.width); got != c.want {
			t.Errorf("truncateLeft(%q, %d) = %q, want %q", c.text, c.width, got, c.want)
		}
	}
}

func TestPaneTitleShowsPathNotName(t *testing.T) {
	s := &chatSession{path: "/tmp/some/deep/project", title: "fix bug"}
	if got := paneTitle(s, 80); got != "/tmp/some/deep/project · fix bug" {
		t.Errorf("wide: %q", got)
	}
	got := paneTitle(s, 30)
	if !strings.HasPrefix(got, "…") || !strings.HasSuffix(got, "project · fix bug") {
		t.Errorf("narrow: %q", got)
	}
}
