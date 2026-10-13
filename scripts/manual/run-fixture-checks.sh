#!/bin/sh
set -eu
root=${1:-/tmp/jin-manual-fixture}
port=${FIXTURE_PORT:-8765}
export FIXTURE_PORT=$port FIXTURE_LOG="$root/artifacts/fixture.jsonl"
python3 "$(dirname "$0")/fake_provider.py" > "$root/artifacts/provider.log" 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true' EXIT INT TERM
for n in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS "http://127.0.0.1:$port/health" >/dev/null; then break; fi
  sleep 0.2
done
python3 - "$port" <<'PY'
import http.client, json, sys
port = int(sys.argv[1])
for path in ("/v1/responses", "/v1/chat/completions", "/v1/messages"):
    c = http.client.HTTPConnection("127.0.0.1", port, timeout=3)
    c.request("POST", path, json.dumps({"model": "synthetic-model", "stream": True}), {"Content-Type": "application/json", "Authorization": "Bearer synthetic-key", "x-api-key": "synthetic-key"})
    r = c.getresponse(); body = r.read().decode()
    assert r.status == 200 and "Synthetic answer." in body, (path, r.status, body)
    c.close()
print("PASS: synthetic SSE responses for Responses, Chat Completions, and Anthropic")
PY
