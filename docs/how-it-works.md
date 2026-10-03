# How jin works

## Layout of the code

```
main.go               wiring
internal/core         agent loop, system prompt, compact, handoff
internal/provider     OpenAI-compatible and Anthropic-compatible streaming client
internal/tools        read, write, edit, bash, ask_user, todo
internal/headless     jin -p, jin models, jin refresh-models (see headless.md)
internal/store        SQLite sessions and settings
internal/hooks        hook files
internal/prompts      reusable prompt files
internal/ui           terminal interface (tcell)
```

UI, core and provider are separate layers. The UI talks to the agent through channels.

## The agent loop

1. You send a message. Jin appends it to the history.
2. The history is streamed to the model together with the tool schemas.
3. If the answer has tool calls, jin runs them, appends the results and calls the model
   again. This repeats until the model answers without tool calls.
4. Messages you type while the agent works are queued and added after the current
   tools finish.
5. `Ctrl+C` interrupts the running turn. Unfinished tool calls get an "interrupted"
   result so the history stays valid.

Every message is saved to the database as it happens (except in `jin -p --no-session`,
which saves nothing).

Provider errors that usually pass on their own are retried up to 5 attempts: HTTP 429,
500, 502, 503, 504, 529, dropped connections and streams cut short. The wait doubles
from 1 s (a `Retry-After` header wins, at most 60 s), and the chat shows a line such as
`Provider returned HTTP 503, retrying in 2s (2/5)`. A stream that sends nothing for 90 s
is cancelled and tried once more; change the limit with the `provider.stall_timeout`
setting (seconds, see [database.md](database.md)). Other errors, such as 401 or 400,
stop the turn at once.

## Tools

- `read`: read a file, optionally with `offset` (1-based line) and `limit`. Pictures
  (png, jpeg, gif, webp, bmp) are attached so the model can see them.
- `write`: create a file or overwrite it.
- `edit`: replace an exact text match in a file.
- `bash`: run a shell command in the working directory. Default timeout 120 s; a command
  still running then moves to the background. Output over 16 KB keeps its first and last
  8 KB; a note says how many bytes and lines were cut and where the full log is
  (`~/.jin/async/<id>.log`, removed after a week), so the model can `read` it instead of
  running the command again.
- `ask_user`: ask the user several questions, each with optional answer options and a
  free-text field. Blocks the turn until answered. Not available in headless mode
  (`jin -p` always removes it).
- `todo`: the model sends the whole list on every call, a call replaces the list. A call
  without `items` reads it. The list is stored in the database per session.

`bash` waits for a command until its timeout (120 s by default) or until you write to the
agent; a command that is still running then moves to the background (see
[async.md](async.md)) instead of being killed.

`edit` and `write` refuse to change a file that changed on disk (size or modification
time) since the agent last read or wrote it in this session; the error tells the model
to read the file again. Your own edits made while the agent works are never lost.

`/tools` switches each tool on or off (setting `tools.disabled`). It applies to
new sessions. With every tool off, no `tools` field is sent to the provider.

Pasting an image saves it under `~/.jin/.pasted` and types its path. Ask the agent to
read that path.

## System prompt

Built when a session starts, in the background: the session is on screen at once, its
input is closed until the texts are ready, and then its agent starts. The system text comes
from `~/.jin/system-prompt.md` or the built-in default (see
[prompts-and-hooks.md](prompts-and-hooks.md)). `{{commands}}` in that text, in hooks and in
`#prompts` run once at this point; nothing else is run. Jin then joins the parts in Go, in
this order: the system text, the tool list, the enabled hooks, the jin docs pointer (always),
the `jin async` instructions with `Your session id` (only with the `bash` tool and a
session id), and the `AGENTS.md` block. The `AGENTS.md` text is never changed or run.

Edits to the system prompt file, hooks, prompts and `AGENTS.md` apply to new sessions only.

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
- **Handoff** (`/handoff`): the model writes a brief and jin opens a new session with that brief in
  the input, ready to edit.

## Data directory

Release builds use `~/.jin`, source builds (`make build`) use `~/.jin-dev`.

```
~/.jin/jin.db        sessions, messages, settings (SQLite, WAL)
~/.jin/AGENTS.md     global context
~/.jin/prompts/      reusable prompts (#plan, #review, #subagents are created at TUI start if missing)
~/.jin/hooks/        hooks
~/.jin/.pasted/      pasted images
```

Files are created with `0600` permissions. Several jin processes can run at once and
share the database. The provider API key is stored in it as plain text.

## Providers and models

`/provider` lists the saved providers. `a` adds one: a kind, a name, a base URL and an API
key. `Enter` makes a provider the active one and asks for its model; `d` deletes one.
Two kinds exist:

- **OpenAI-compatible**: `POST {base_url}/chat/completions`, models from `{base_url}/models`.
  Base URL example: `https://api.openai.com/v1`.
- **Anthropic-compatible**: `POST {base_url}/v1/messages` with the `x-api-key` header, models
  from `{base_url}/v1/models`. Base URL example: `https://api.anthropic.com`.

A session keeps the provider it started with. Switching the active provider opens a new
session when the current one already has messages. Each provider has its own model list
cache and its own scope.

`/model` picks the model and reasoning effort. `/scope` limits which models appear in the
picker and the `Ctrl+M` rotation. Effort is remembered per model.

**Reasoning effort.** jin asks an OpenAI-compatible provider which levels a model accepts
(it sends a deliberately invalid level and reads the answer). When that fails, or the answer
cannot be read, it offers `low`, `medium` and `high`. If the provider accepts the invalid
level, the model has only the default. The Anthropic kind has no such probe and always offers
`low`, `medium` and `high`. A chosen level is sent as `thinking: {"type": "adaptive"}` with
`output_config: {"effort": "<level>"}`; `Default` sends neither. If the provider answers
HTTP 400 to that, jin repeats the request once without them and does not send them again for
that model during this run.

## Settings commands

`/model`, `/scope`, `/provider`, `/sound`, `/tools` and `/change-editor`.

- **Sound**: Toggle, When (always or on blur), Volume. macOS plays a system sound,
  other systems get the terminal bell.
- **Tools**: On/Off per tool. New sessions only.
- **Editor**: nano, vim or hx, used for editing prompts, hooks, `AGENTS.md` and the input.
