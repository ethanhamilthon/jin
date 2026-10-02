package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestMarkdownLinksCarryTheirURL(t *testing.T) {
	rows := markdownRows("see [docs](https://x.dev/a) and <https://y.dev>", 60)
	if got := linkAt(rows[0], 6); got != "https://x.dev/a" {
		t.Fatalf("link at docs = %q", got)
	}
	if got := linkAt(rows[0], 2); got != "" {
		t.Fatalf("plain text has link %q", got)
	}
	found := false
	for x := 0; x < 60; x++ {
		found = found || linkAt(rows[0], x) == "https://y.dev"
	}
	if !found {
		t.Fatal("autolink lost its URL")
	}
}

func TestClickOnLinkOpensIt(t *testing.T) {
	a, _ := layoutApp(t)
	var opened string
	linkOpener = func(link string) { opened = link }
	t.Cleanup(func() { linkOpener = openLink })
	s := a.active
	s.rows = markdownRows("[docs](https://x.dev)", 60)
	s.view = viewport{first: 0, height: 10}
	s.handleMouse(tcell.NewEventMouse(3, 0, tcell.ButtonPrimary, tcell.ModNone), a.screen)
	s.handleMouse(tcell.NewEventMouse(3, 0, tcell.ButtonNone, tcell.ModNone), a.screen)
	if opened != "https://x.dev" {
		t.Fatalf("opened %q", opened)
	}
	if s.selection.active {
		t.Fatal("a click on a link should not leave a selection")
	}
}
