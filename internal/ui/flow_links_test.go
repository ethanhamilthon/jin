package ui

import (
	"testing"

	"jin/internal/core"
)

func TestAnswerLinksListsTheLastAnswer(t *testing.T) {
	s := &chatSession{width: 80}
	s.appendEntry(chatEntry{kind: core.UpdateAssistant, text: "old https://old.example"})
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "more"})
	s.appendEntry(chatEntry{kind: core.UpdateAssistant, text: "See [the docs](https://go.dev/doc) and https://pkg.go.dev, again [docs](https://go.dev/doc)."})
	links := s.answerLinks()
	if len(links) != 2 || links[0].value != "https://go.dev/doc" || links[0].label != "the docs" || links[1].value != "https://pkg.go.dev" {
		t.Fatalf("links = %+v", links)
	}
	if links[1].detail != "" {
		t.Errorf("bare URL should not repeat itself: %+v", links[1])
	}
}

func TestLinksFlowOpensTheChoice(t *testing.T) {
	a, _ := layoutApp(t)
	var opened string
	old := linkOpener
	linkOpener = func(link string) { opened = link }
	t.Cleanup(func() { linkOpener = old })
	a.active.appendEntry(chatEntry{kind: core.UpdateAssistant, text: "[a](https://a.example) [b](https://b.example)"})
	a.openLinksFlow()
	a.sel.move(1)
	a.submitSelector()
	if opened != "https://b.example" {
		t.Fatalf("opened %q", opened)
	}
}
