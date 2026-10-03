package ui

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

const resultGutter = "│ "

// toolResultRows draws the lines of resultText under their tool call, each
// clipped to one row: output in grey, removed lines red, added lines green.
func toolResultRows(text string, width int) []chatRow {
	var rows []chatRow
	room := max(1, width-2-len(resultGutter)-1)
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		mark, body := resultMark(line)
		style := resultStyle(mark)
		spans := []chatSpan{{text: resultGutter, style: border}}
		if mark == '-' || mark == '+' || mark == hiddenMark {
			spans = append(spans, chatSpan{text: string(mark), style: style})
		} else {
			spans = append(spans, chatSpan{text: " ", style: style})
		}
		spans = append(spans, chatSpan{text: truncate(expandTabs(body), room), style: style})
		rows = append(rows, chatRow{kind: core.UpdateToolResult, spans: spans})
	}
	return rows
}

func resultMark(line string) (rune, string) {
	if strings.HasPrefix(line, string(hiddenMark)) {
		return hiddenMark, " " + strings.TrimPrefix(line, string(hiddenMark))
	}
	return rune(line[0]), line[1:]
}

func resultStyle(mark rune) tcell.Style {
	switch mark {
	case '-':
		return base.Foreground(colorRed)
	case '+':
		return base.Foreground(colorGreen)
	case hiddenMark:
		return dim.Italic(true)
	}
	return muted
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)

// expandTabs makes a line of program output safe to draw: colors are
// dropped, tabs become spaces and other control characters spaces too.
func expandTabs(text string) string {
	text = strings.ReplaceAll(ansiEscape.ReplaceAllString(text, ""), "\t", "    ")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
}
