package files

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
)

func Resolve(raw, home, cwd string) (string, bool) {
	p := pathFor(raw, home, cwd)
	st, err := os.Stat(p)
	if err != nil || !st.Mode().IsRegular() {
		return "", false
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", false
	}
	return filepath.Clean(abs), true
}

func Block(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<attached-files>\n")
	for _, p := range paths {
		b.WriteString(`<file path="`)
		b.WriteString(xmlEscape(p))
		b.WriteString(`"/>`)
		b.WriteByte('\n')
	}
	b.WriteString("</attached-files>")
	return b.String()
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func Extract(text, home, cwd string) (string, []string) {
	var out strings.Builder
	seen := map[string]bool{}
	var paths []string
	for pos := 0; pos < len(text); {
		at := strings.IndexByte(text[pos:], '@')
		if at < 0 {
			out.WriteString(text[pos:])
			break
		}
		t, ok := Find(text, pos+at+1)
		if !ok {
			out.WriteString(text[pos : pos+at+1])
			pos += at + 1
			continue
		}
		out.WriteString(text[pos:t.Start])
		if p, valid := Resolve(t.Path, home, cwd); valid {
			out.WriteString(p)
			if !seen[p] {
				seen[p] = true
				paths = append(paths, p)
			}
		} else {
			out.WriteString(text[t.Start:t.End])
		}
		pos = t.End
		if pos == len(text) {
			break
		}
	}
	return out.String(), paths
}
