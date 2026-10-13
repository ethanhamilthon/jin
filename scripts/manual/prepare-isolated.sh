#!/bin/sh
set -eu
root=${1:-/tmp/jin-manual-fixture}
mkdir -p "$root/home" "$root/project" "$root/data" "$root/artifacts"
printf '%s\n' 'synthetic-only fixture home' > "$root/home/README.txt"
printf '%s\n' 'synthetic manual test project' > "$root/project/README.md"
chmod 700 "$root/home" "$root/data"
printf '%s\n' "$root"
