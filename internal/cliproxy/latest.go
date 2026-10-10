package cliproxy

import (
	"context"
	"strconv"
	"strings"
)

// LatestVersion is the newest compatible release on GitHub, or DefaultVersion
// when the list cannot be read.
func LatestVersion(ctx context.Context) string {
	versions, err := Versions(ctx)
	if err != nil {
		return DefaultVersion
	}
	best, bestPatch := DefaultVersion, patchOf(DefaultVersion)
	for _, version := range versions {
		if patch := patchOf(version); patch > bestPatch {
			best, bestPatch = version, patch
		}
	}
	return best
}

func patchOf(version string) int {
	patch, _ := strconv.Atoi(strings.TrimPrefix(version, "v8.0."))
	return patch
}

// InstallLatest installs the newest compatible release when none is installed.
func InstallLatest(ctx context.Context, root string) error {
	if Installed(root).Version != "" {
		return nil
	}
	return InstallInitial(ctx, root, LatestVersion(ctx))
}
