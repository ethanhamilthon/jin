package ui

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var imageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".bmp": true}

// imageLink turns the destination of a Markdown image into a link that
// openLink accepts: web addresses stay, paths become file: URLs.
func imageLink(dest string) string {
	if u, err := url.Parse(dest); err == nil && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "file") {
		return dest
	}
	if rest, ok := strings.CutPrefix(dest, "~/"); ok {
		home, _ := os.UserHomeDir()
		dest = filepath.Join(home, rest)
	}
	abs, err := filepath.Abs(dest)
	if err != nil {
		return ""
	}
	return (&url.URL{Scheme: "file", Path: abs}).String()
}

// isImageFile reports whether a file: link names an existing picture, the
// only local files that are opened from model text.
func isImageFile(u *url.URL) bool {
	info, err := os.Stat(u.Path)
	return u.Scheme == "file" && err == nil && info.Mode().IsRegular() && imageExts[strings.ToLower(filepath.Ext(u.Path))]
}
