package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		args []string
		want Kind
	}{
		{nil, TUI},
		{[]string{"-p", "hi"}, Headless},
		{[]string{"--model", "x", "-p", "hi"}, Headless},
		{[]string{"models"}, Headless},
		{[]string{"refresh-models"}, Headless},
		{[]string{"async", "run", "ls"}, Unknown},
		{[]string{"export", "abc"}, Export},
		{[]string{"hooks", "list"}, Hooks},
		{[]string{"--help"}, Help},
		{[]string{"-h"}, Help},
		{[]string{"help"}, Help},
		{[]string{"--version"}, Version},
		{[]string{"version"}, Version},
		{[]string{"update"}, Update},
		{[]string{"web", "--port", "8000"}, Web},
		{[]string{"sessions"}, Sessions},
		{[]string{"sessions", "list"}, Sessions},
		{[]string{"sessions", "search", "word"}, Sessions},
		{[]string{"sessions", "compact", "abc"}, Headless},
		{[]string{"sessions", "reload", "abc"}, Headless},
		{[]string{"foo"}, Unknown},
		{[]string{"--bogus"}, Unknown},
		{[]string{"hello", "--", "-p"}, Unknown},
	}
	for _, c := range cases {
		if got := Classify(c.args); got != c.want {
			t.Errorf("Classify(%v) = %v, want %v", c.args, got, c.want)
		}
	}
}

func TestMessages(t *testing.T) {
	var out bytes.Buffer
	PrintUnknown(&out, []string{"foo", "bar"})
	if got := out.String(); !strings.Contains(got, `unknown command "foo"`) || !strings.Contains(got, "jin --help") {
		t.Errorf("unknown = %q", got)
	}
	out.Reset()
	PrintVersion(&out, "v0.4")
	if out.String() != "jin v0.4\n" {
		t.Errorf("version = %q", out.String())
	}
	out.Reset()
	PrintHelp(&out)
	for _, want := range []string{"jin export", "jin models", "jin -p"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help misses %q", want)
		}
	}
}
