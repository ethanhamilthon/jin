#!/bin/sh
set -eu
root=${1:-/tmp/jin-manual-fixture}
dir=$(cd "$(dirname "$0")" && pwd)
export PYTHONDONTWRITEBYTECODE=1
mkdir -p "$root/artifacts"
python3 -W error -m unittest discover -s "$dir" -p 'test_fixture.py' 2>&1 | tee "$root/artifacts/fixture-unit.log"
echo "PASS: fixture unit checks (streams, rejection, title dispatch, delay)"
