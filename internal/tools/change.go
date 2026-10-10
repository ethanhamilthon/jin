package tools

import (
	"context"
	"errors"
	"os"
)

// Change is one file write by edit or write: the content before it, so the
// UI can show a diff.
type Change struct {
	Path string
	// Existed is false when the write created the file.
	Existed bool
	Before  string
	After   string
}

type changeSinkKey struct{}

// WithChangeSink makes edit and write report every file they change to fn.
func WithChangeSink(ctx context.Context, fn func(Change)) context.Context {
	return context.WithValue(ctx, changeSinkKey{}, fn)
}

func reportChange(ctx context.Context, change Change) {
	if fn, ok := ctx.Value(changeSinkKey{}).(func(Change)); ok {
		fn(change)
	}
}

// currentContent reads a file that is about to be overwritten. A missing file
// is not an error: it is reported as not existing.
func currentContent(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}
