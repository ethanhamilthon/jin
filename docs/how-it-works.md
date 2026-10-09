# How jin works

## Layout of the code

```
main.go               wiring
internal/core         agent loop, system prompt, compact, handoff
internal/agentkit     builds the agent the same way for the TUI, jin web and jin -p
internal/provider     Chat Completions, Responses and Anthropic streaming clients, retries
internal/tools        read, write, edit, bash, task, ask_user
internal/diff         line diffs for the tool output fold
internal/cli          which command a command line means
internal/headless     jin -p, jin models, jin refresh-models (see headless.md)
internal/export       jin export
internal/tasks        background tasks owned by the jin process
internal/store        SQLite project registry, sessions and settings
internal/hooks        hook files, project hooks, jin hooks
internal/prompts      reusable prompt files
internal/dyn          {{commands}} in prompts and hooks
internal/sysprompt    ~/.jin/system-prompt.md
internal/startup      builds a session's prompts before its agent starts
internal/upgrade      one-time steps when a new version starts
internal/update       jin update and the background release check
internal/datadir      reset and swap data operations
internal/ui           terminal interface (tcell)
internal/session      live sessions for any front end, history entries shared with the TUI
internal/web          jin web: HTTP API, Server-Sent Events, embedded UI (see web.md)
web/                  the jin web UI (Svelte), built into internal/web/dist
```

UI, core and provider are separate layers. The UI talks to the agent through channels;
jin web reaches the same agent through `internal/session`.
Development rules are in `AGENTS.md` at the repository root.

### Build, test, release

Requires Go 1.27 or newer.

```sh
make build        # bin/jin with the web UI, uses ~/.jin-dev for data
make build-prod   # bin/jin, uses ~/.jin for data
make install      # build-prod and copy to /usr/local/bin
make check        # go test + go vet
make golden       # rewrite the TUI screen snapshots after a change to the look
```

Pushing a tag (`git tag v0.8.0 && git push origin v0.8.0`) runs
`.github/workflows/release.yml` on a macOS runner: `make check`, then
`scripts/build-release.sh <tag>`, which
writes `dist/jin_<os>_<arch>.tar.gz` and `dist/checksums.txt` and attaches them to a GitHub
release. Without a local clone, run the release workflow by hand (Actions → release → Run
workflow) with the tag name: it releases `main` and creates the tag. `make release
VERSION=v0.8.0` builds the same archives locally.

## The agent loop

1. You send a message. Jin appends it to the history.
2. The history is streamed to the model together with the tool schemas.
3. If the answer has tool calls, jin runs them, appends the results and calls the model
   again. This repeats until the model answers without tool calls. The calls of one answer
   run at the same time, `bash` included; results keep the order of the calls. `ask_user`
   runs alone, and a call on a file that an earlier call of the same answer
   changes (or reads, when this call changes it) waits for that call.
4. Messages you type while the agent works are queued and added after the current
   tools finish.
5. `Ctrl+C` interrupts the running turn. Unfinished tool calls get an "interrupted"
   result so the history stays valid.

Every message is saved to the database as it happens (except in `jin -p --no-session`,
which saves nothing).

Provider errors that usually pass on their own are retried up to 5 attempts: HTTP 429,
500, 502, 503, 504, 529, dropped connections and streams cut short. The wait doubles
from 1 s (a `Retry-After` header wins, at most 60 s), and the chat shows a line such as
`Provider returned HTTP 503, retrying in 2s (2/5)`. A stream that stays silent too long
is cancelled and tried again. The limit depends on the reasoning effort: 90 s by default
and for `none` or `minimal`, 120 s for `low`, 180 s for `medium`, 300 s for `high` and
600 s for `xhigh` or `max`. Every received byte restarts the timer, and `Ctrl+C` cancels
at once. The `provider.stall_timeout` setting (seconds, see [database.md](database.md))
replaces this policy with one fixed limit. Other errors, such as 401 or 400, stop the turn
at once.

## Providers

A provider has one of three kinds:

- `openai`: OpenAI-compatible `POST /chat/completions`.
- `responses`: OpenAI `POST /responses`. Jin sends the full history itself, always with
  `store: false`, and asks for `reasoning.encrypted_content`, so reasoning survives
  between turns without server-side state.
- `anthropic`: Anthropic-compatible `POST /v1/messages`, with automatic prompt caching
  (`cache_control`).

Output that the chat format cannot hold (encrypted reasoning items, thinking blocks with
their signatures) is saved with the message as `jin_native` and sent back verbatim to the
same provider kind and model. A different model or kind gets the plain text and tool
calls instead. Chat Completions requests never carry `jin_native`.

Usage counts all input tokens, cached or not. Cache reads and writes are shown only when
the provider reports them; a missing field is not treated as zero.

### Debug log

`JIN_DEBUG=1` writes a provider log to `~/.jin/debug/provider-<pid>.jsonl`
(`~/.jin-dev` for source builds); `JIN_DEBUG=/path/file.jsonl` picks the file. It is off
by default. Each line is one JSON event: `attempt_start` (effort and silence limit),
`request` (sizes, hashes and how much of the history prefix is unchanged since the last
request), `http_headers`, `first_byte`, `raw_usage`, `attempt_end` (normalized usage and
cache hit ratio), `retry`, `stall` (`first_byte` or `stream` phase) and compaction
events. The log holds metadata only: no prompts, file contents, tool arguments, headers
or API keys. The file is created with mode 0600.

## Tools

- `read`: read a file, optionally with `offset` and `limit` (both 1-based, at least 1).
  Pictures (png, jpeg, gif, webp, bmp) are attached so the model can see them; an image
  over 40 megapixels or a file over 20 MB is refused with an error that names the limit.
- `write`: create a file or overwrite it.
- `edit`: replace an exact text match in a file. When the text is not found but matches
  after ignoring line endings, trailing whitespace or tabs, the error names the cause and
  the line; the match is never applied automatically. `old_string` equal to `new_string`
  is an error.
- `bash`: run a shell command in the working directory, or in `dir` (relative to it or
  absolute) when given. `cd` does not carry over to the next call; `dir` is the way to run
  in a subdirectory. Default timeout 120 s; a command
  still running then moves to the background (in headless mode, `jin -p`, it is killed).
  Each call starts a new shell with no TTY and no stdin. A non-zero exit ends the output with
  `[exit code: N]`, not a failure note: for `grep`, `diff` or `test` it is an answer. Output over 16 KB keeps its first and last
  8 KB; a note says how many bytes and lines were cut and where the full log is
  (`~/.jin/tasks/<id>.log`, removed after a week), so the model can `read` it instead of
  running the command again.
- `task`: background tasks: `start` a command and get its id at once, `check`, `input`,
  `stop`, `list`. Results arrive as messages by themselves. See [tasks.md](tasks.md).
- `ask_user`: ask the user several questions, each with optional answer options and a
  free-text field. Blocks the turn until answered. Not available in headless mode
  (`jin -p` always removes it).

`bash` runs every command in its own session (`Setsid`), so `sudo`, `ssh` and editors
cannot take over the terminal of jin, and sets `GIT_TERMINAL_PROMPT=0`, `GIT_EDITOR=true`,
`GIT_PAGER=cat`, `PAGER=cat` and `DEBIAN_FRONTEND=noninteractive`. The `!` shell in the
TUI gets the same setup, and cancelling it ends the whole process group.

`bash` waits for a command until its timeout (120 s by default) or until you write to the
agent; a command that is still running then moves to the background (see
[tasks.md](tasks.md)) instead of being killed.

`edit` and `write` refuse to change a file that changed on disk (size or modification
time) since the agent last read or wrote it in this session, whoever changed it (you or a
command such as `sed -i`, a formatter or `git checkout`); the error tells the model to
read the file again. Files are tracked by their real path, so a symlink and its target
count as one file. Your own edits made while the agent works are never lost.

The Tools row of `/settings` switches each tool on or off (setting `tools.disabled`). It
applies to new sessions. With every tool off, no `tools` field is sent to the provider.

Pasting an image (`Ctrl+V`) saves it under `~/.jin/.pasted` and types its path. On macOS
the clipboard is read with `osascript` and `pbpaste`; on Linux with `wl-paste` (Wayland)
or `xclip` (X11) when installed. Without them `Ctrl+V` pastes nothing, but the terminal's
own paste still works for text. Copying a selection uses `pbcopy`, `wl-copy` or `xclip`,
and falls back to the terminal (OSC 52). Ask the agent to
read that path.

## System prompt

Built when a session starts, in the background: the session is on screen at once, its
input is closed until the texts are ready, and then its agent starts. The system text comes
from `~/.jin/system-prompt.md` or the built-in default (see
[prompts-and-hooks.md](prompts-and-hooks.md)). `{{commands}}` in that text, in hooks and in
`#prompts` run once at this point; nothing else is run. Jin then joins the parts in Go, in
this order, most stable first: the system text, a `Tools: none` line only when no tool is on (the tool schemas carry the
tool descriptions), the jin docs pointer (always),
the enabled hooks
and the `AGENTS.md` block. A cache break line follows, then the environment (working
directory, OS, date) and `Your session id`. The `AGENTS.md` text is never changed or run.

The break line is for the provider adapters, which never send it. Anthropic gets two system
blocks with a cache breakpoint on the first, so a new session in the same project still
reuses the cached start of the prompt. Chat Completions and Responses get both parts as one
text.

In the TUI, the prompt is built again when a session opens, before the next request after
compaction or handoff, or before a request after more than 5 minutes idle when the prompt
is over an hour old or from a previous day. During active turns, jin leaves the prompt
unchanged. When the new text differs,
the next user message starts with `<system-refreshed>instructions were refreshed</system-refreshed>`.
Edits to the system prompt file, hooks and `AGENTS.md` apply at these points. `#prompts` are
rendered only when a session opens.

The default system text asks the model to keep the user's goal and constraints across turns
and compaction, continue until completion or a clear blocker, and review the result before
finishing. That review covers relevant edge cases and regressions, fixes critical issues
within scope, and runs available checks when possible. Jin does not enforce a separate
validation step. A custom `# system` section replaces these default instructions.

## Context files

Every `AGENTS.md` that applies is included:

1. The global one: `~/.jin/AGENTS.md`.
2. Ones in parent directories, marked "not the current project".
3. The one in the working directory.

The start screen lists the files used. Edit them in your own editor; jin does not edit them.

## Compact and handoff

- **Compact** (`/compact`): the model summarizes the session. The chat shows a
  divider and the model continues from the summary. The full history stays in the
  database. It also runs on its own at 80% of the model's context window (windows come
  from the OpenRouter and LiteLLM catalogues). The status bar shows the context as
  `◫ 42K/200K 21%` and turns amber from 70%. An automatic run is announced in the chat,
  and its divider says how much it saved: `Auto-compacted 162K → 4.1K tokens`.
- **Prune**: from 70% of the window, jin replaces the output of tool calls older than
  the last 4 turns with `[output of read omitted, 31200 bytes; run it again if needed]`.
  This needs no model call and keeps the conversation itself. Only the history sent to the
  model is pruned; the database keeps the full text. Compaction runs only when the context
  is still at 80% after pruning.
- **Context size**: jin takes the size the provider reported for the last response and
  adds about one token per 4 bytes for every message added since, such as tool results.
  After a compaction the size is estimated from the new history and the tool schemas.
- **Unknown window**: for a model the catalogues do not know (local models, proxies),
  set its window in tokens, and automatic compaction works for it too:
  `sqlite3 ~/.jin/jin.db "INSERT OR REPLACE INTO settings VALUES ('models.window.<model id>', '32768')"`.
- **Context overflow**: when the provider refuses a request as too long, jin shows one
  notice, prunes, compacts when pruning is not enough, and retries the request once. A
  second refusal is shown as an error. When a compaction or handoff request is itself too
  long, it is sent again with all old tool output left out; only the last 2 turns keep theirs.
- **Handoff** (`/handoff`): the model writes a brief and jin opens a new session with that brief in
  the input, ready to edit.

## Data directory

Release builds use `~/.jin`, source builds (`make build`) use `~/.jin-dev`.

```
~/.jin/jin.db        projects, sessions, messages, settings (SQLite, WAL)
~/.jin/AGENTS.md     global context
~/.jin/prompts/      your reusable prompts (#plan and #review are built into the binary)
~/.jin/hooks/        hooks
~/.jin/.pasted/      pasted images
```

Files are created with `0600` permissions. Several jin processes can run at once and
share the database. The provider API key is stored in it as plain text.

## Providers and models

`/provider` lists saved providers. Add one with its kind, name, base URL and API
key and choose the default for new sessions. Changing the default never reroutes a live
session. Three provider kinds exist:

- **OpenAI Chat Completions**: `POST {base_url}/chat/completions`, models from `{base_url}/models`.
  Base URL example: `https://api.openai.com/v1`.
- **OpenAI Responses**: `POST {base_url}/responses`, models from `{base_url}/models`.
  Base URL example: `https://api.openai.com/v1`.
- **Anthropic**: `POST {base_url}/v1/messages` with the `x-api-key` header, models
  from `{base_url}/v1/models`. Base URL example: `https://api.anthropic.com`.

A session keeps the provider it started with. Changing the default provider affects new
sessions; existing sessions continue with their saved provider. A session whose provider was
deleted is cut off ("Provider <id> of this session was deleted") until you pick a new default
with `/provider`, which moves those sessions to it. Each provider has its own model list cache
and scope.

`/model` picks the focused session's model and reasoning effort. Each row shows what the
OpenRouter and LiteLLM catalogues know: context window, input and output price per 1M tokens,
and `reasoning` / `vision` support, for example `200K ctx · $1.25 / $10 · reasoning`.
`/provider` sets which models appear in the picker and `Ctrl+M` rotation. The Scoped models
row of `/settings` chooses the rotation. Effort is remembered per model.

**Reasoning effort.** jin asks an OpenAI-compatible provider which levels a model accepts
(it sends a deliberately invalid level and reads the answer). When that fails, or the answer
cannot be read, it offers `low`, `medium` and `high`. If the provider accepts the invalid
level, the model has only the default. The Anthropic kind has no such probe and always offers
`low`, `medium` and `high`. A chosen level is sent as `thinking: {"type": "adaptive"}` with
`output_config: {"effort": "<level>"}`; `Default` sends neither. If the provider answers
HTTP 400 to that, jin repeats the request once without them and does not send them again for
that model during this run.

## Global settings and session actions

The global settings live in `/settings`: its rows are Sound (Audio), Swap config and Reset
(Data), Change editor (Editor), Tools (Tools), Scoped models (Models), Motion (Appearance),
Prompts, Hooks and System prompt (Instructions). `/theme` stays its own command. They are
global, not session-local. The global Tools setting
applies to new sessions. System instructions load for new sessions and `/reload`
rebuilds the focused session's prompt. The editor setting selects nano, vim or hx.

- **Sound**: Toggle, When (always or on blur), Volume. macOS plays a system sound; other
  systems get the terminal bell.
