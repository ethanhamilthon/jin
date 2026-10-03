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
		{[]string{"async", "run", "ls"}, Async},
		{[]string{"daemon"}, Daemon},
		{[]string{"export", "abc"}, Export},
		{[]string{"--help"}, Help},
		{[]string{"-h"}, Help},
		{[]string{"help"}, Help},
		{[]string{"--version"}, Version},
		{[]string{"version"}, Version},
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

func TestNeedsDB(t *testing.T) {
	for _, kind := range []Kind{Help, Version, Unknown} {
		if NeedsDB(kind) {
			t.Errorf("kind %v must not open the database", kind)
		}
	}
	for _, kind := range []Kind{TUI, Headless, Async, Daemon} {
		if !NeedsDB(kind) {
			t.Errorf("kind %v needs the database", kind)
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
	for _, want := range []string{"jin async run", "jin models", "jin -p"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help misses %q", want)
		}
	}
}
