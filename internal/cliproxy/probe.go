package cliproxy

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func probeBinary(ctx context.Context, root, version string) error {
	temporary, err := os.MkdirTemp(root, ".probe-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	k, err := loadKeys(temporary)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(temporary, "versions", version), 0o700); err != nil {
		return err
	}
	if err = os.Symlink(binaryPath(root, version), binaryPath(temporary, version)); err != nil {
		return err
	}
	p, err := startProcess(ctx, temporary, version, k)
	if err != nil {
		return err
	}
	defer p.stop()
	raw, err := p.management(ctx, "GET", "/force-model-prefix", nil)
	if err != nil {
		return err
	}
	var flags struct {
		Force bool `json:"force-model-prefix"`
	}
	if err = json.Unmarshal(raw, &flags); err != nil || !flags.Force {
		return errors.New("proxy does not enforce source-prefixed models")
	}
	return nil
}
