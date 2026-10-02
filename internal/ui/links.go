package ui

import (
	"net/url"
	"os/exec"
	"runtime"

	"github.com/clipperhouse/displaywidth"
)

// linkAt is the URL of the link drawn at screen column col of row, if any.
func linkAt(row chatRow, col int) string {
	x := 2
	for _, span := range row.spans {
		w := displaywidth.String(span.text)
		if col >= x && col < x+w {
			_, link := span.style.GetUrl()
			return link
		}
		x += w
	}
	return ""
}

// linkOpener opens a clicked link; tests swap it out.
var linkOpener = openLink

// openLink hands a web or mail link to the system's opener. Other schemes,
// such as file: or custom app handlers, are never opened from model text.
func openLink(link string) {
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "mailto") {
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", link)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", link)
	default:
		cmd = exec.Command("xdg-open", link)
	}
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}
