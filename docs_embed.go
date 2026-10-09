package main

import (
	"embed"
	"io/fs"
)

//go:embed docs/*.md
var docPages embed.FS

// docsFS is the documentation built into the binary, pages at its root.
func docsFS() fs.FS {
	pages, _ := fs.Sub(docPages, "docs")
	return pages
}
