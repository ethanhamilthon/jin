//go:build !unix

package async

import (
	"context"

	"jin/internal/store"
)

// Serve is only built for unix systems; Windows support comes later.
func Serve(context.Context, *store.DB, string) error { return errUnsupported }

func startDetached(string) error { return errUnsupported }
