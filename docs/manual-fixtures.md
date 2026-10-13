# Manual integration fixture evidence

Use this fixture only with a disposable build and synthetic data. It does not edit product packages or installed Jin data.

## Start the fixture

1. Create isolated paths with `sh scripts/manual/prepare-isolated.sh`.
2. Start `FIXTURE_MODE=delay-first FIXTURE_LOG=/tmp/jin-manual-fixture/artifacts/provider.jsonl python3 scripts/manual/fake_provider.py`.
3. Configure the disposable Jin build with base URL `http://127.0.0.1:8765/v1`, key `synthetic-key`, model `synthetic-model`, and the requested protocol type.
4. Keep the app, data folder, and project inside `/tmp/jin-manual-fixture`. Use a free port and do not stop other servers.

The server supports Responses (`/responses`), Chat Completions (`/chat/completions`), and Anthropic (`/messages`). `FIXTURE_MODE` accepts `ok`, `delay-first`, `title-failure`, `unsupported`, `bare400`, `bare404`, `auth`, `quota`, `timeout`, `partial`, and `tool-loop`. Set `FIXTURE_DELAY` for wait duration. All keys and model names are synthetic.

## Manual test execution

| Case | Exact action | Expected result | Evidence artifact |
| --- | --- | --- | --- |
| Protocol round-trip | Run `sh scripts/manual/run-fixture-checks.sh /tmp/jin-manual-fixture` | Three HTTP 200 SSE responses with synthetic answer text | `artifacts/provider.log`, `artifacts/fixture.jsonl` |
| TUI interaction | Run `HOME=/tmp/jin-manual-fixture/home JIN_DATA_DIR=/tmp/jin-manual-fixture/data expect scripts/manual/tui-send.exp <disposable-jin> 'synthetic question' /tmp/jin-manual-fixture/artifacts/tui.log` | TUI receives the fixture answer | `artifacts/tui.log` |
| Browser interaction | Start Jin web on a free port. Open `http://localhost:5173/api/state?token=<printed-token>`, then `http://localhost:5173/` with `agent-browser --session jin-web-manual --args '--no-browser-gpu,--disable-gpu,--use-angle=swiftshader' open <url>` | Web UI loads and sends synthetic prompt | `artifacts/browser-snapshot.txt`, screenshot, fixture request log |
| Headless interaction | Run isolated `jin -p --no-session --cwd /tmp/jin-manual-fixture/project --model synthetic-model --provider <fixture-provider-id> 'synthetic question'` | Command prints synthetic answer without UI | `artifacts/headless.log`, fixture request log |
| Early title | Use `delay-first`, set `FIXTURE_DELAY=8`, send first prompt in a new saved interactive session | A separate title request arrives before the delayed final answer | `artifacts/fixture.jsonl` timestamps and UI/TUI capture |
| Title failure | Restart fixture with `FIXTURE_MODE=title-failure`, then send first prompt | Main answer continues; title notice appears; excerpt title stays | `artifacts/title-failure.log`, fixture request log |
| Protocol rejection | Restart with `unsupported`, send through Responses endpoint | One explicit unsupported-protocol response permits one alternate request | `artifacts/unsupported.jsonl` |
| Bare errors | Repeat with `bare400`, `bare404`, `auth`, and `quota` | No protocol switch for status-only 400/404, auth, or quota | One request per test in `artifacts/fixture.jsonl` |
| Timeout and partial stream | Repeat with `timeout` then `partial` | No alternate protocol after timeout or partial output | Request log plus client error capture |
| Tool question/answer/error tones | Use `tool-loop`; accept the `ask_user` question, answer it, then cause a tool error in the disposable project | Question, answer, and error sound events trigger in UI/TUI preferences | `artifacts/sound-actions.txt` and UI/TUI capture |

Add the exact command, local date, observed result, and artifact paths to the run record. Keep synthetic credentials out of screenshots and logs. Do not call a harness check an integration pass. Do not call browser audio decoding proof of audible human quality.

## Evidence status

| Area | Status | Limit |
| --- | --- | --- |
| Fixture HTTP/SSE checks | Runnable locally | These validate the fixture protocol bodies only |
| Actual browser/TUI/headless manual checks | Pending integrator build | No integrated product UI pass is claimed here |
| Audio HTTP decode/play checks | Not run | A browser audio API check cannot prove a listener hears the tone |
| Audible tone quality | Unverified | Requires a human listener |
| Native Wails/macOS launch | Blocked | CPU-only WKWebView mode is not established |
| Linux runtime | Excluded | No Linux runtime tests are requested |
