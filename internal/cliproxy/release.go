package cliproxy

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const DefaultVersion = "v8.0.23"
const releaseBase = "https://github.com/router-for-me/CLIProxyAPI/releases/download/"
const maxBinary = 256 << 20

var compatibleVersion = regexp.MustCompile(`^v8\.0\.[0-9]+$`)

func ValidVersion(version string) bool {
	if !compatibleVersion.MatchString(version) {
		return false
	}
	patch, err := strconv.Atoi(strings.TrimPrefix(version, "v8.0."))
	return err == nil && patch >= 23
}
func asset(version, os, arch string) (string, error) {
	if !ValidVersion(version) {
		return "", fmt.Errorf("CLIProxyAPI version must match the supported v8.0.x contract")
	}
	if os != "darwin" && os != "linux" {
		return "", fmt.Errorf("CLIProxyAPI is supported on macOS/Linux; use WSL2 on Windows")
	}
	if arch == "arm64" {
		arch = "aarch64"
	}
	if arch != "aarch64" && arch != "amd64" {
		return "", fmt.Errorf("unsupported CLIProxyAPI architecture")
	}
	return "CLIProxyAPI_" + version[1:] + "_" + os + "_" + arch + ".tar.gz", nil
}
func binaryPath(root, version string) string {
	return filepath.Join(root, "versions", version, "cli-proxy-api")
}
func currentAsset(version string) (string, error) {
	return asset(version, runtime.GOOS, runtime.GOARCH)
}
