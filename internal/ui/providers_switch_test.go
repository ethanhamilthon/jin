package ui

import (
	"strings"
	"testing"
)

func TestProviderListShowsOnOffSwitches(t *testing.T) {
	a, screen := layoutApp(t)
	sel := a.openList("Providers", []option{{label: "Work", detail: "Anthropic · url", value: "w", choices: []string{"On", "Off"}}}, "w", func(string) error { return nil })
	sel.twoLines = true
	a.draw()
	for y := 0; y < 24; y++ {
		if text := rowText(screen, y, 60); strings.Contains(text, "Work") {
			if !strings.Contains(text, "On") || !strings.Contains(text, "Off") {
				t.Fatalf("row %q does not show the switch", text)
			}
			return
		}
	}
	t.Fatal("provider row not drawn")
}
