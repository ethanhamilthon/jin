# Manual integration fixture evidence

Use this fixture only with a disposable build and synthetic data. It does not edit product packages or installed Jin data.

## Start the fixture

1. Create isolated paths with `sh scripts/manual/prepare-isolated.sh`.
2. Start a task (not a foreground command): `FIXTURE_MODE=delay-first FIXTURE_PORT=8765 FIXTURE_LOG=/tmp/jin-manual-fixture/artifacts/fixture.jsonl python3 scripts/manual/fake_provider.py`. Pick another `FIXTURE_PORT` if 8765 is busy. The server binds to 127.0.0.1 only.
3. Configure the disposable Jin build with key `synthetic-key`, model `synthetic-model`, and the requested protocol type. Base URL for OpenAI Chat and Responses: `http://127.0.0.1:8765/v1`. Base URL for Anthropic: `http://127.0.0.1:8765` (Jin adds `/v1/messages`).
4. Isolation: Jin stores data in `$HOME/.jin-dev` (source build) or `$HOME/.jin`. Set `HOME=/tmp/jin-manual-fixture/home` for every Jin process. There is no `JIN_DATA_DIR` variable. For `jin -p` you can also set `JIN_BASE_URL`, `JIN_API_KEY`, and `JIN_PROVIDER_KIND`.
5. Keep the project inside `/tmp/jin-manual-fixture`. Do not stop other servers. Stop the fixture only with `task stop`.

The server supports Responses (`/responses`), Chat Completions (`/chat/completions`), and Anthropic (`/messages`). The server reads the last message of each request. If it contains `title` (change with `FIXTURE_TITLE_MARKER`), the request is a title request. Each request is logged as `phase:request` with `kind` (`main` or `title`) and `seq` (count per kind). A `phase:respond` line is logged after the stream is sent. A title request answers `Synthetic title`. A main request answers `Synthetic answer.`. `GET /v1/models` lists `synthetic-model`.

`FIXTURE_MODE` accepts `ok`, `delay-first` (the first main request waits `FIXTURE_DELAY` seconds, title requests do not), `title-failure` (title requests get HTTP 500), `unsupported` (HTTP 400 with code `unsupported_endpoint` on `/responses` only), `unsupported-chat` (the same on `/chat/completions` only), `bare400`, `bare404`, `auth`, `quota`, `timeout`, `partial`, and `tool-loop`. Set `FIXTURE_DELAY` for wait duration. All keys and model names are synthetic.

## Manual test execution

| Case | Exact action | Expected result | Evidence artifact |
| --- | --- | --- | --- |
| Protocol round-trip | Run `sh scripts/manual/run-fixture-checks.sh /tmp/jin-manual-fixture` | Unit checks pass: valid streams for all three protocols, rejection modes, title dispatch, delayed first answer | `artifacts/fixture-unit.log` |
| TUI interaction | Run `HOME=/tmp/jin-manual-fixture/home TERM=xterm-256color expect scripts/manual/tui-send.exp <disposable-jin> 'synthetic question' /tmp/jin-manual-fixture/artifacts/tui.log` | TUI receives the fixture answer | `artifacts/tui.log` |
| Browser interaction | Start Jin web on a free port. Open `http://localhost:5173/api/state?token=<printed-token>`, then `http://localhost:5173/` with `agent-browser --session jin-web-manual --args '--no-browser-gpu,--disable-gpu,--use-angle=swiftshader' open <url>` | Web UI loads and sends synthetic prompt | `artifacts/browser-snapshot.txt`, screenshot, fixture request log |
| Headless interaction | Run `HOME=/tmp/jin-manual-fixture/home jin -p --no-session --cwd /tmp/jin-manual-fixture/project --model synthetic-model --provider <fixture-provider-id> 'synthetic question'` | Command prints synthetic answer without UI | `artifacts/headless.log`, fixture request log |
| Early title | Use `delay-first`, set `FIXTURE_DELAY=8`, send first prompt in a new saved interactive session | A separate title request arrives before the delayed final answer | `artifacts/fixture.jsonl`: the `title` request line has an earlier `time` than the `respond` line of main `seq` 1; and UI/TUI capture |
| Title failure | Restart fixture with `FIXTURE_MODE=title-failure`, then send first prompt | Main answer continues; title notice appears; excerpt title stays | `artifacts/title-failure.log`, fixture request log |
| Protocol rejection | Restart with `unsupported`, send through the Responses transport. Use `unsupported-chat` for the reverse | One explicit unsupported-protocol response permits one alternate request | `artifacts/fixture.jsonl` |
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
