package cliproxy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Installation struct {
	Version  string `json:"version"`
	Previous string `json:"previous,omitempty"`
}

func Installed(root string) Installation {
	var state Installation
	_ = readJSON(filepath.Join(root, "installed.json"), &state)
	if !ValidVersion(state.Version) {
		return Installation{}
	}
	if _, err := os.Stat(binaryPath(root, state.Version)); err != nil {
		return Installation{}
	}
	return state
}

func Install(ctx context.Context, root, version string) error {
	name, err := currentAsset(version)
	if err != nil {
		return err
	}
	archive, err := download(ctx, releaseBase+version+"/"+name, maxBinary)
	if err != nil {
		return err
	}
	sums, err := download(ctx, releaseBase+version+"/checksums.txt", 1<<20)
	if err != nil {
		return err
	}
	if err = verify(archive, sums, name); err != nil {
		return err
	}
	binary, err := extract(archive)
	if err != nil {
		return err
	}
	if err = writeFile(binaryPath(root, version), binary, 0o700); err != nil {
		return err
	}
	return nil
}

func InstallInitial(ctx context.Context, root, version string) error {
	unlock, err := lock(root, "install.lock")
	if err != nil {
		return err
	}
	defer unlock()
	if Installed(root).Version != "" {
		return errors.New("CLIProxyAPI is installed; use Update")
	}
	if err = Install(ctx, root, version); err != nil {
		return err
	}
	if err = checkBinary(ctx, root, version); err != nil {
		return err
	}
	return writeJSON(filepath.Join(root, "installed.json"), Installation{Version: version})
}

func checkBinary(ctx context.Context, root, version string) error {
	probe, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return probeBinary(probe, root, version)
}
