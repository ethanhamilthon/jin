#!/bin/sh
# Builds the desktop app for one platform into dist/.
# Usage: scripts/build-desktop.sh <version> <os>/<arch>   (for example v0.15.1 darwin/arm64)
# macOS gives jin-desktop_darwin_<arch>.dmg with an ad-hoc signed Jin.app.
# Linux gives jin-desktop_linux_<arch>.tar.gz and needs GTK 3 and WebKitGTK 4.1.
# The web UI must be built first (make web). The build needs a C toolchain.
set -eu

version="${1:?usage: build-desktop.sh <version> <os>/<arch>}"
target="${2:?usage: build-desktop.sh <version> <os>/<arch>}"
os="${target%/*}"
arch="${target#*/}"
cd "$(dirname "$0")/.."
test -f internal/web/dist/index.html || { echo "build the web UI first: make web" >&2; exit 1; }
mkdir -p dist
flags="-s -w -X jin/internal/paths.mode=prod -X main.version=${version}"
tags="desktop,production"

case "$os" in
darwin)
	app="dist/app-${arch}/Jin.app"
	rm -rf "dist/app-${arch}"
	mkdir -p "${app}/Contents/MacOS"
	CGO_ENABLED=1 GOOS=darwin GOARCH="$arch" \
		CGO_LDFLAGS="-framework UniformTypeIdentifiers -mmacosx-version-min=10.13" \
		CGO_CFLAGS="-mmacosx-version-min=10.13" \
		go build -trimpath -tags "$tags" -ldflags "$flags" -o "${app}/Contents/MacOS/Jin" ./cmd/jin-desktop
	mkdir -p "${app}/Contents/Resources" "dist/AppIcon.iconset"
	for px in 16 32 128 256 512; do
		sips -z "$px" "$px" build/icon/icon-1024.png --out "dist/AppIcon.iconset/icon_${px}x${px}.png" >/dev/null
		sips -z "$((px * 2))" "$((px * 2))" build/icon/icon-1024.png --out "dist/AppIcon.iconset/icon_${px}x${px}@2x.png" >/dev/null
	done
	iconutil -c icns "dist/AppIcon.iconset" -o "${app}/Contents/Resources/AppIcon.icns"
	rm -rf "dist/AppIcon.iconset"
	short="${version#v}"
	cat > "${app}/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>Jin</string>
	<key>CFBundleDisplayName</key><string>Jin</string>
	<key>CFBundleIdentifier</key><string>dev.jin.desktop</string>
	<key>CFBundleExecutable</key><string>Jin</string>
	<key>CFBundleIconFile</key><string>AppIcon</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>${short}</string>
	<key>CFBundleVersion</key><string>${short}</string>
	<key>LSMinimumSystemVersion</key><string>10.13</string>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST
	codesign --force --deep -s - "$app"
	ln -sf /Applications "dist/app-${arch}/Applications"
	rm -f "dist/jin-desktop_darwin_${arch}.dmg"
	hdiutil create -quiet -volname "Jin ${version}" -srcfolder "dist/app-${arch}" -ov -format UDZO "dist/jin-desktop_darwin_${arch}.dmg"
	rm -rf "dist/app-${arch}"
	echo "built dist/jin-desktop_darwin_${arch}.dmg"
	;;
linux)
	name="jin-desktop_linux_${arch}"
	rm -rf "dist/${name}"
	mkdir -p "dist/${name}"
	CGO_ENABLED=1 GOOS=linux GOARCH="$arch" \
		go build -trimpath -tags "${tags},webkit2_41" -ldflags "$flags" -o "dist/${name}/jin-desktop" ./cmd/jin-desktop
	tar -C "dist/${name}" -czf "dist/${name}.tar.gz" jin-desktop
	rm -rf "dist/${name}"
	echo "built dist/${name}.tar.gz"
	;;
*)
	echo "unsupported target ${target}" >&2
	exit 1
	;;
esac
