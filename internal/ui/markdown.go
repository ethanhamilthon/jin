package ui

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var mdParser = goldmark.New(goldmark.WithExtensions(extension.GFM))

// markdownRows parses text into a CommonMark+GFM AST and renders it to
// wrapped chatRow/chatSpan rows, supporting headings, nested lists,
// blockquotes, fenced code blocks, tables, task lists, and inline styles.
func markdownRows(raw string, width int) []chatRow {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	src := []byte(raw)
	doc := mdParser.Parser().Parse(text.NewReader(src))
	return renderBlockChildren(doc, src, width, 0)
}

func renderBlockChildren(parent ast.Node, source []byte, width, depth int) []chatRow {
	var rows []chatRow
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		rows = append(rows, renderBlock(c, source, width, depth)...)
		if c.NextSibling() != nil && needsBlankSeparator(c) {
			rows = append(rows, chatRow{})
		}
	}
	return rows
}

func needsBlankSeparator(n ast.Node) bool {
	switch n.(type) {
	case *ast.Paragraph, *ast.Heading, *ast.List, *ast.Blockquote, *ast.FencedCodeBlock, *ast.CodeBlock, *east.Table:
		return true
	}
	return false
}

func renderBlock(n ast.Node, source []byte, width, depth int) []chatRow {
	switch v := n.(type) {
	case *ast.Heading:
		spans := inlineSpans(n, source, headingStyle(v.Level))
		return wrapMarkdown(spans, width)
	case *ast.Paragraph:
		spans := inlineSpans(n, source, bodyStyle)
		return wrapMarkdown(spans, width)
	case *ast.List:
		return renderList(v, source, width, depth)
	case *ast.Blockquote:
		return renderBlockquote(v, source, width, depth)
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		return renderCodeBlock(n, source, width)
	case *east.Table:
		return renderTable(v, source, width)
	case *ast.ThematicBreak:
		rule := strings.Repeat("─", max(1, width))
		return []chatRow{{text: rule, spans: []chatSpan{{text: rule, style: border}}}}
	default:
		return renderBlockChildren(n, source, width, depth)
	}
}
