package ui

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// clipboardTool is a command line program that reads the clipboard, found on
// PATH. Wayland's wl-paste wins over xclip when a Wayland session runs.
type clipboardTool struct {
	listTypes []string
	readImage func(mime string) []string
	readText  []string
}

func linuxClipboard() (clipboardTool, bool) {
	if os.Getenv("WAYLAND_DISPLAY") != "" && onPath("wl-paste") {
		return clipboardTool{
			listTypes: []string{"wl-paste", "--list-types"},
			readImage: func(mime string) []string { return []string{"wl-paste", "--no-newline", "--type", mime} },
			readText:  []string{"wl-paste", "--no-newline"},
		}, true
	}
	if os.Getenv("DISPLAY") != "" && onPath("xclip") {
		return clipboardTool{
			listTypes: []string{"xclip", "-selection", "clipboard", "-t", "TARGETS", "-o"},
			readImage: func(mime string) []string { return []string{"xclip", "-selection", "clipboard", "-t", mime, "-o"} },
			readText:  []string{"xclip", "-selection", "clipboard", "-o"},
		}, true
	}
	return clipboardTool{}, false
}

func onPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func output(args []string) ([]byte, error) {
	return exec.Command(args[0], args[1:]...).Output()
}

// imageTypes are the picture formats asked for, best first.
var imageTypes = []string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp"}

// linuxImage reads a picture from the clipboard. It is saved as it comes;
// the read tool tells the format from the bytes, not the file name.
func (c clipboardTool) linuxImage() ([]byte, bool) {
	types, err := output(c.listTypes)
	if err != nil {
		return nil, false
	}
	offered := strings.Fields(string(types))
	for _, mime := range imageTypes {
		for _, t := range offered {
			if t != mime {
				continue
			}
			if data, err := output(c.readImage(mime)); err == nil && len(data) > 0 {
				return data, true
			}
		}
	}
	return nil, false
}

// clipboardText reads plain text from the system clipboard.
func clipboardText() (string, bool) {
	args := []string{"pbpaste"}
	if runtime.GOOS != "darwin" {
		tool, ok := linuxClipboard()
		if !ok {
			return "", false
		}
		args = tool.readText
	}
	out, err := output(args)
	if err != nil {
		return "", false
	}
	return string(bytes.TrimSuffix(out, []byte{0})), true
}

// clipboardImage reads a picture from the system clipboard, if it holds one.
func clipboardImage() ([]byte, bool) {
	if runtime.GOOS == "darwin" {
		return clipboardImagePNG()
	}
	if tool, ok := linuxClipboard(); ok {
		return tool.linuxImage()
	}
	return nil, false
}
