package diff

import (
	"strings"
	"testing"
)

func render(lines []Line) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteByte(byte(l.Op))
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

func TestLines(t *testing.T) {
	cases := []struct{ before, after, want string }{
		{"a\nb\nc\n", "a\nB\nc\n", "-b\n+B\n"},
		{"a\nc", "a\nb\nc", "+b\n"},
		{"", "x\ny\n", "+x\n+y\n"},
		{"x\ny\nz", "y", "-x\n-z\n"},
		{"same", "same", ""},
	}
	for _, c := range cases {
		if got := render(Lines(c.before, c.after)); got != c.want {
			t.Errorf("Lines(%q, %q) = %q, want %q", c.before, c.after, got, c.want)
		}
	}
}
