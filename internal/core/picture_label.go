package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"path/filepath"

	"jin/internal/provider"
)

// PictureLabel is the chat note of a picture a tool returned, for example
// "[shot.png · image/png · 84 KB · 1920×1080]". Without a path it has no name.
func PictureLabel(path string, picture provider.Image) string {
	label := picture.Label()
	if raw, err := base64.StdEncoding.DecodeString(picture.Data); err == nil {
		if config, _, err := image.DecodeConfig(bytes.NewReader(raw)); err == nil {
			label += fmt.Sprintf(" · %d×%d", config.Width, config.Height)
		}
	}
	if path != "" {
		label = filepath.Base(path) + " · " + label
	}
	return "[" + label + "]"
}

// CallPath is the path argument of a tool call, or "" when it has none.
func CallPath(call provider.ToolCall) string {
	var args struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal([]byte(call.Function.Arguments), &args)
	return args.Path
}
