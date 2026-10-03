package ui

import (
	"strings"
	"testing"
)

func TestWrapChatKeepsWordsWhole(t *testing.T) {
	cases := []struct {
		text  string
		width int
		want  []string
	}{
		{"hello brave new world", 11, []string{"hello brave", "new world"}},
		{"hello brave new world", 8, []string{"hello", "brave", "new", "world"}},
		{"abcdefghij xy", 4, []string{"abcd", "efgh", "ij", "xy"}},
		{"a  b", 2, []string{"a", "b"}},
		{"line one\nline two", 20, []string{"line one", "line two"}},
		{"привет мир кириллица", 10, []string{"привет мир", "кириллица"}},
		{"日本語 テキスト", 8, []string{"日本語", "テキスト"}},
		{"", 5, []string{""}},
	}
	for _, c := range cases {
		got := wrapChat(c.text, c.width)
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("wrapChat(%q, %d) = %q, want %q", c.text, c.width, got, c.want)
		}
	}
}

func TestWrapRowsFitWidth(t *testing.T) {
	text := strings.Repeat("lorem ipsum dolor sit amet, consectetur adipiscing ", 20)
	for width := 1; width < 40; width++ {
		var joined strings.Builder
		for _, row := range wrapChat(text, width) {
			if n := len([]rune(row)); n > width {
				t.Fatalf("width %d: row %q has %d cells", width, row, n)
			}
			joined.WriteString(row)
		}
		if strings.ReplaceAll(joined.String(), " ", "") != strings.ReplaceAll(text, " ", "") {
			t.Fatalf("width %d lost text", width)
		}
	}
}

func TestWrapMarkdownKeepsWordsAndStyles(t *testing.T) {
	spans := []chatSpan{{text: "plain words ", style: base}, {text: "bold words", style: accent}}
	rows := wrapMarkdown(spans, 13)
	var texts []string
	for _, r := range rows {
		texts = append(texts, r.text)
	}
	if strings.Join(texts, "|") != "plain words|bold words" {
		t.Fatalf("rows = %q", texts)
	}
	if len(rows[1].spans) != 1 || rows[1].spans[0].style != accent {
		t.Fatalf("styles = %+v", rows[1].spans)
	}
}

func TestInputWrapsByWordsAndKeepsCursor(t *testing.T) {
	input := clusters("hello brave world")
	lines, row, col := wrapInput(input, len(input), 12)
	var got []string
	total := 0
	for _, line := range lines {
		got = append(got, strings.Join(line, ""))
		total += len(line)
	}
	if strings.Join(got, "|") != "hello brave |world" || total != len(input) {
		t.Fatalf("lines = %q", got)
	}
	if row != 1 || col != 5 {
		t.Fatalf("cursor = %d,%d", row, col)
	}
}
