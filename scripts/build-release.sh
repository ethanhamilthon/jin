#!/bin/sh
# Builds release archives for every supported platform into dist/.
# Usage: scripts/build-release.sh <version>   (for example v0.1)
set -eu

version="${1:?usage: build-release.sh <version>}"
cd "$(dirname "$0")/.."
rm -rf dist
mkdir dist

for target in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do
	os="${target%/*}"
	arch="${target#*/}"
	name="jin_${os}_${arch}"
	echo "building ${name}"
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath \
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
