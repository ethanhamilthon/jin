package web

import "testing"

func TestSafeNameDropsDirectories(t *testing.T) {
	for in, want := range map[string]string{"../../etc/passwd": "passwd", "a\\b.txt": "b.txt", "": "file", "..": "..", "report.pdf": "report.pdf"} {
		if got := safeName(in); got != want {
			t.Errorf("safeName(%q) = %q, want %q", in, got, want)
		}
	}
}
