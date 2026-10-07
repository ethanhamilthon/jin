package ui

import (
	"net/url"
	"os"
	"path/filepath"
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

func TestImageLinkOpensOnlyPictures(t *testing.T) {
	dir := t.TempDir()
	picture, text := filepath.Join(dir, "a.png"), filepath.Join(dir, "a.txt")
	for _, path := range []string{picture, text} {
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for link, want := range map[string]bool{imageLink(picture): true, imageLink(text): false, imageLink(filepath.Join(dir, "missing.png")): false} {
		u, _ := url.Parse(link)
		if isImageFile(u) != want {
			t.Fatalf("%s: want %v", link, want)
		}
	}
	if got := imageLink("https://x.test/a.png"); got != "https://x.test/a.png" {
		t.Fatal(got)
	}
}
