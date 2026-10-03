#!/bin/sh
# Installs jin from the latest GitHub release.
#   curl -fsSL https://raw.githubusercontent.com/ethanhamilthon/jin/main/install.sh | sh
# Run it again to update: an installed jin is replaced in place.
# Options (environment): JIN_VERSION=v0.6.2 to pin a release, JIN_INSTALL_DIR to
# choose the target directory (default: the directory of the installed jin,
# else /usr/local/bin, or ~/.local/bin when that is not writable).
set -eu

REPO="ethanhamilthon/jin"

fail() {
	echo "jin install: $*" >&2
	exit 1
}

case "$(uname -s)" in
Darwin) os=darwin ;;
Linux) os=linux ;;
*) fail "unsupported system $(uname -s); jin supports macOS and Linux" ;;
esac

case "$(uname -m)" in
arm64 | aarch64) arch=arm64 ;;
x86_64 | amd64) arch=amd64 ;;
*) fail "unsupported architecture $(uname -m)" ;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

archive="jin_${os}_${arch}.tar.gz"
if [ -n "${JIN_VERSION:-}" ]; then
	base="https://github.com/${REPO}/releases/download/${JIN_VERSION}"
else
	base="https://github.com/${REPO}/releases/latest/download"
fi

current="$(command -v jin 2>/dev/null || true)"
dir="${JIN_INSTALL_DIR:-}"
if [ -z "$dir" ] && [ -n "$current" ] && [ -w "$(dirname "$current")" ]; then
	dir="$(dirname "$current")"
fi
if [ -z "$dir" ]; then
	dir=/usr/local/bin
	[ -w "$dir" ] || dir="$HOME/.local/bin"
fi
previous=""
if [ -x "${dir}/jin" ]; then
	previous="$("${dir}/jin" --version 2>/dev/null | awk '{print $2}' || true)"
fi
mkdir -p "$dir" || fail "cannot create $dir"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "downloading ${archive}"
curl -fsSL "${base}/${archive}" -o "${tmp}/${archive}" || fail "download failed: ${base}/${archive}"
curl -fsSL "${base}/checksums.txt" -o "${tmp}/checksums.txt" || fail "cannot download checksums.txt"

expected="$(grep " ${archive}\$" "${tmp}/checksums.txt" | cut -d ' ' -f 1)"
[ -n "$expected" ] || fail "no checksum listed for ${archive}"
if command -v sha256sum >/dev/null 2>&1; then
	actual="$(sha256sum "${tmp}/${archive}" | cut -d ' ' -f 1)"
else
	actual="$(shasum -a 256 "${tmp}/${archive}" | cut -d ' ' -f 1)"
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for ${archive}"

tar -xzf "${tmp}/${archive}" -C "$tmp" jin
install -m 755 "${tmp}/jin" "${dir}/jin"
installed="$("${dir}/jin" --version 2>/dev/null | awk '{print $2}' || true)"
if [ -z "$previous" ]; then
	echo "installed jin ${installed} to ${dir}/jin"
elif [ "$previous" = "$installed" ]; then
	echo "jin ${installed} is already the latest; reinstalled ${dir}/jin"
else
	echo "updated jin ${previous} -> ${installed} (${dir}/jin)"
fi
if [ -n "$current" ] && [ "$current" != "${dir}/jin" ]; then
	echo "note: another jin comes first on your PATH: ${current}"
fi

case ":${PATH}:" in
*":${dir}:"*) ;;
*) echo "add ${dir} to your PATH:  export PATH=\"${dir}:\$PATH\"" ;;
esac
echo "run jin from a project directory to start"
