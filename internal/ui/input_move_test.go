package ui

import "testing"

func TestMoveVertical(t *testing.T) {
	input := clusters("abcd\nef\nghijk")
	cases := []struct{ cursor, delta, want int }{
		{3, 1, 7},
		{7, 1, 10},
		{10, -1, 7},
		{7, -1, 2},
		{2, -1, 0},
		{12, 1, 13},
	}
	for _, c := range cases {
		if got := moveVertical(input, c.cursor, 20, c.delta); got != c.want {
			t.Errorf("moveVertical(cursor %d, delta %d) = %d, want %d", c.cursor, c.delta, got, c.want)
		}
	}
}

func TestMoveVerticalAcrossWrappedLines(t *testing.T) {
	input := clusters("abcdefgh")
	if got := moveVertical(input, 6, 4, -1); got != 2 {
		t.Errorf("up from the second wrapped line = %d, want 2", got)
	}
	if got := moveVertical(input, 1, 4, 1); got != 5 {
		t.Errorf("down from the first wrapped line = %d, want 5", got)
	}
}

func TestFitScroll(t *testing.T) {
	cases := []struct{ start, cursor, total, height, want int }{
		{0, 2, 10, 4, 0},
		{0, 5, 10, 4, 2},
		{4, 5, 10, 4, 4},
		{4, 2, 10, 4, 2},
		{9, 0, 3, 4, 0},
	}
	for _, c := range cases {
		if got := fitScroll(c.start, c.cursor, c.total, c.height); got != c.want {
			t.Errorf("fitScroll%v = %d, want %d", c, got, c.want)
		}
	}
}
