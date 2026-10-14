// Package icon holds the picture of the desktop app icon.
package icon

import _ "embed"

// PNG is the 1024 pixel icon, drawn from icon.svg by render.swift.
//
//go:embed icon-1024.png
var PNG []byte
