//go:build !unix

package cliproxy

import (
	"errors"
	"path/filepath"
)

func socketPath(root string) string { return filepath.Join(root, "control.sock") }
func socketDirectory(string) error  { return errors.New("managed proxy requires macOS or Linux") }
