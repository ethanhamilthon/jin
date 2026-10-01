package ui

import (
	"github.com/gdamore/tcell/v3"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// inlineSpans walks n's inline children (n must have inline content, e.g. a
// Paragraph, Heading, or TableCell) into styled spans.
func inlineSpans(n ast.Node, source []byte, style tcell.Style) []chatSpan {
	var spans []chatSpan
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		spans = append(spans, renderInline(c, source, style)...)
	}
	return spans
}

func renderInline(n ast.Node, source []byte, style tcell.Style) []chatSpan {
	switch v := n.(type) {
	case *ast.Text:
		text := string(v.Segment.Value(source))
		if v.SoftLineBreak() || v.HardLineBreak() {
			text += " "
		}
		return []chatSpan{{text: text, style: style}}
	case *ast.String:
		return []chatSpan{{text: string(v.Value), style: style}}
	case *ast.CodeSpan:
		return []chatSpan{{text: string(n.Text(source)), style: style.Background(colorRaised)}}
	case *ast.Emphasis:
		next := style.Italic(true)
		if v.Level >= 2 {
			next = style.Bold(true)
		}
		return inlineSpans(n, source, next)
	case *east.Strikethrough:
		return inlineSpans(n, source, style.StrikeThrough(true))
	case *east.TaskCheckBox:
		box := "[ ] "
		if v.IsChecked {
			box = "[x] "
		}
		return []chatSpan{{text: box, style: style}}
	case *ast.AutoLink:
		return []chatSpan{{text: string(v.URL(source)), style: accent.Underline(true)}}
	case *ast.Link:
		label := inlineSpans(n, source, accent.Underline(true))
		return append(label, chatSpan{text: " (" + string(v.Destination) + ")", style: style.Foreground(colorDim)})
	case *ast.Image:
		alt := plainText(inlineSpans(n, source, style))
		return []chatSpan{{text: "[image: " + alt + "]", style: style.Foreground(colorDim)}}
	case *ast.RawHTML:
		return []chatSpan{{text: string(n.Text(source)), style: style.Foreground(colorDim)}}
	default:
		return inlineSpans(n, source, style)
	}
}

func plainText(spans []chatSpan) string {
	var out []byte
	for _, s := range spans {
		out = append(out, s.text...)
	}
	return string(out)
}
