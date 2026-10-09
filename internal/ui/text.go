package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clipperhouse/displaywidth"

	"jin/internal/prompts"
)

// truncateWidth cuts text to at most width cells on grapheme boundaries.
func truncateWidth(text string, width int) string {
	var b strings.Builder
	used := 0
	graphemes := displaywidth.StringGraphemes(text)
	for graphemes.Next() {
		if used+graphemes.Width() > width {
			break
		}
		b.WriteString(graphemes.Value())
		used += graphemes.Width()
	}
	return b.String()
}

func truncate(text string, width int) string {
	if width < 1 {
		return ""
	}
	if displaywidth.String(text) <= width {
		return text
	}
	return truncateWidth(text, width-1) + "…"
}

// truncateLeft keeps the end of text, which is the part of a path that
// tells folders apart.
func truncateLeft(text string, width int) string {
	if width < 1 {
		return ""
	}
	if displaywidth.String(text) <= width {
		return text
	}
	var parts []string
	graphemes := displaywidth.StringGraphemes(text)
	for graphemes.Next() {
		parts = append(parts, graphemes.Value())
	}
	used, from := 1, len(parts)
	for from > 0 {
		w := displaywidth.String(parts[from-1])
		if used+w > width {
			break
		}
		used += w
		from--
	}
	return "…" + strings.Join(parts[from:], "")
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func shortPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && (path == home || strings.HasPrefix(path, home+string(os.PathSeparator))) {
		return "~" + path[len(home):]
	}
	return path
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

// filePreview is the start of a markdown file's text on one line.
func filePreview(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return prompts.Summary(string(data))
}
