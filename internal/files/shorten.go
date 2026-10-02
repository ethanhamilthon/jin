package files

import (
	"path/filepath"
	"strings"

	"github.com/clipperhouse/displaywidth"
)

func Shorten(path string, width int) string {
	if width <= 0 {
		return ""
	}
	if displaywidth.String(path) <= width {
		return path
	}
	name := filepath.Base(path)
	if displaywidth.String(name) >= width {
		return trimWidth(name, width)
	}
	avail := width - displaywidth.String(name) - 2
	dir := filepath.Dir(path)
	return trimWidth(dir, avail) + "…" + string(filepath.Separator) + name
}

func trimWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if displaywidth.String(s) <= width {
		return s
	}
	left := (width - 1) / 2
	right := width - 1 - left
	var a, b strings.Builder
	for _, r := range s {
		x := string(r)
		if displaywidth.String(a.String()+x) > left {
			break
		}
		a.WriteRune(r)
	}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		x := string(runes[i])
		if displaywidth.String(x+b.String()) > right {
			break
		}
		b.WriteRune(runes[i])
	}
	return a.String() + "…" + reverse(b.String())
}
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
