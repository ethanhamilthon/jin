package ui

import (
	"testing"

	"jin/internal/tools"
)

func TestGoldenPanels(t *testing.T) {
	a, screen := goldenApp(t)
	conversation(a.active)
	a.openList("Theme", []option{{label: "Jin Original", value: "a"}, {label: "Nord", detail: "cool", value: "b"}}, "b", func(string) error { return nil })
	a.draw()
	golden(t, screen, "panel-list")

	a, screen = goldenApp(t)
	a.active.ask = newAskState([]tools.Question{{Question: "Which mode?", Options: []string{"fast", "safe"}}})
	a.draw()
	golden(t, screen, "panel-ask")

	a, screen = goldenApp(t)
	a.active.ask = newAskState([]tools.Question{{Question: "Which of these approaches should the agent take for the long-line wrapping change?", Options: []string{"Wrap between words and split only words longer than a whole row", "Keep the old grapheme wrapping"}}})
	a.draw()
	golden(t, screen, "panel-ask-long")
}
