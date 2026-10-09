package docs

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

var pages = fstest.MapFS{
	"README.md":   {Data: []byte("# Index\n")},
	"headless.md": {Data: []byte("# Headless\n")},
	"notes.txt":   {Data: []byte("not a page")},
}

func run(args ...string) (string, string, int) {
	var out, errOut bytes.Buffer
	code := Main(args, pages, &out, &errOut)
	return out.String(), errOut.String(), code
}

func TestPointerListAndPage(t *testing.T) {
	if out, _, code := run(); code != 0 || !strings.Contains(out, "jin docs README") {
		t.Errorf("pointer: %q %d", out, code)
	}
	if out, _, _ := run("--list"); out != "README\nheadless\n" {
		t.Errorf("list: %q", out)
	}
	for _, arg := range []string{"headless", "headless.md"} {
		if out, _, code := run(arg); out != "# Headless\n" || code != 0 {
			t.Errorf("page %s: %q %d", arg, out, code)
		}
	}
}

func TestUnknownPageAndBadFlag(t *testing.T) {
	if _, errOut, code := run("nope"); code != 1 || !strings.Contains(errOut, "README, headless") {
		t.Errorf("unknown page: %q %d", errOut, code)
	}
	if _, _, code := run("--what"); code != 2 {
		t.Errorf("bad flag exit %d", code)
	}
	if _, _, code := run("../go.mod"); code != 1 {
		t.Errorf("a path outside the pages must not be read: exit %d", code)
	}
}
