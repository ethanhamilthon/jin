package ui

import "github.com/yuin/goldmark/ast"

func renderBlockquote(bq *ast.Blockquote, source []byte, width, depth int) []chatRow {
	bar := chatSpan{text: "│ ", style: dim}
	inner := renderBlockChildren(bq, source, max(1, width-spansWidth([]chatSpan{bar})), depth)
	out := make([]chatRow, len(inner))
	for i, row := range inner {
		if row.text == "" && len(row.spans) == 0 {
			out[i] = chatRow{text: "│", spans: []chatSpan{{text: "│", style: dim}}}
			continue
		}
		out[i] = prependRow([]chatSpan{bar}, row)
	}
	return out
}

func renderCodeBlock(n ast.Node, source []byte, width int) []chatRow {
	raw := string(n.Lines().Value(source))
	lines := splitLines(raw)
	var rows []chatRow
	for _, line := range lines {
		rows = append(rows, wrapMarkdown([]chatSpan{{text: line, style: muted}}, width)...)
	}
	return rows
}

func splitLines(s string) []string {
	for len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	return append(lines, s[start:])
}
