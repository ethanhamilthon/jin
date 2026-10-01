package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
	east "github.com/yuin/goldmark/extension/ast"
)

func padCell(spans []chatSpan, targetWidth int, align east.Alignment) []chatSpan {
	current := spansWidth(spans)
	if current > targetWidth {
		return truncateCell(spans, targetWidth)
	}
	pad := targetWidth - current
	leftPad, rightPad := 0, pad
	switch align {
	case east.AlignRight:
		leftPad, rightPad = pad, 0
	case east.AlignCenter:
		leftPad = pad / 2
		rightPad = pad - leftPad
	}
	var out []chatSpan
	if leftPad > 0 {
		out = append(out, chatSpan{text: strings.Repeat(" ", leftPad)})
	}
	out = append(out, spans...)
	if rightPad > 0 {
		out = append(out, chatSpan{text: strings.Repeat(" ", rightPad)})
	}
	return out
}

func truncateCell(spans []chatSpan, targetWidth int) []chatSpan {
	if targetWidth <= 1 {
		return []chatSpan{{text: "…"[:min(1, targetWidth)]}}
	}
	budget := targetWidth - 1
	var out []chatSpan
	used := 0
	for _, s := range spans {
		sw := displaywidth.String(s.text)
		if used+sw <= budget {
			out = append(out, s)
			used += sw
			continue
		}
		for _, r := range s.text {
			rw := displaywidth.String(string(r))
			if used+rw > budget {
				break
			}
			out = append(out, chatSpan{text: string(r), style: s.style})
			used += rw
		}
		break
	}
	out = append(out, chatSpan{text: "…", style: border})
	return out
}

func renderTableBorder(colWidths []int) chatRow {
	var spans []chatSpan
	spans = append(spans, chatSpan{text: "├", style: border})
	for i, w := range colWidths {
		if i > 0 {
			spans = append(spans, chatSpan{text: "┼", style: border})
		}
		spans = append(spans, chatSpan{text: strings.Repeat("─", w+2), style: border})
	}
	spans = append(spans, chatSpan{text: "┤", style: border})
	var text strings.Builder
	for _, s := range spans {
		text.WriteString(s.text)
	}
	return chatRow{text: text.String(), spans: spans}
}
