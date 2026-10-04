#!/bin/sh
# Builds release archives for every supported platform into dist/.
# Usage: scripts/build-release.sh <version>   (for example v0.7.2)
# The microphone of /voice needs cgo: run this on a Mac with the Xcode command
# line tools for the darwin builds, and have zig on PATH for the linux builds.
set -eu

version="${1:?usage: build-release.sh <version>}"
cd "$(dirname "$0")/.."
rm -rf dist
mkdir dist

command -v zig >/dev/null 2>&1 || { echo "zig is required for the linux builds" >&2; exit 1; }

for target in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do
	os="${target%/*}"
	arch="${target#*/}"
	name="jin_${os}_${arch}"
	echo "building ${name}"
	case "$os/$arch" in
	linux/amd64) cc="zig cc -target x86_64-linux-gnu" ;;
	linux/arm64) cc="zig cc -target aarch64-linux-gnu" ;;
	darwin/amd64) cc="clang -arch x86_64" ;;
	darwin/arm64) cc="clang -arch arm64" ;;
	esac
	CGO_ENABLED=1 CC="$cc" GOOS="$os" GOARCH="$arch" go build -trimpath \
		-ldflags "-s -w -X jin/internal/paths.mode=prod -X main.version=${version}" \
		-o "dist/jin" .
	tar -C dist -czf "dist/${name}.tar.gz" jin
	rm dist/jin
done

cd dist
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum ./*.tar.gz | sed 's| \./| |' > checksums.txt
else
	shasum -a 256 ./*.tar.gz | sed 's| \./| |' > checksums.txt
fi
cat checksums.txt
