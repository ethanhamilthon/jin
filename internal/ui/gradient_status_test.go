package ui

import (
	"strings"
	"testing"
)

func TestStatusLineSitsOnBlue(t *testing.T) {
	a, screen := layoutApp(t)
	a.draw()
	_, style, _ := screen.Get(0, 23)
	if style.GetBackground() != colorBlue {
		t.Fatalf("status row background = %v", style.GetBackground())
	}
	_, style, _ = screen.Get(0, 22)
	if style.GetBackground() == colorBlue {
		t.Fatal("the rule above the status must stay black")
	}
	if got := rowText(screen, 21, 60); !strings.HasPrefix(got, "❯") {
		t.Fatalf("input prefix = %q", got)
	}
}
