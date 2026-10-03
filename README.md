# jin

A minimal TUI coding agent written in GO.

- Any OpenAI-compatible or Anthropic-compatible API; save several providers and switch between them
- Input-first TUI: slash commands (`/model`, `/new`, `/bash`, ...) work anywhere in the text
- `@file` mentions with autocomplete, `~/` paths and quoted names
- Streaming / Markdown rendering
- Tools: `read`, `write`, `edit`, `bash`, `ask_user`, `todo` (each can be switched off with `/tools`)
- Images: `read` shows a picture to the model (png, jpeg, gif, webp, bmp). Pasting an image
  saves it to disk and types its path, so ask the agent to read that path
- Several sessions at once, saved per directory in SQLite
- Reusable prompts: type `#name` in the input, manage them with `/prompts`
- Several jin processes can run at once and share one database

## Install

macOS and Linux, arm64 and amd64. The script downloads the latest release, checks its
SHA-256 and installs `jin` into `/usr/local/bin` (or `~/.local/bin` when that is not
writable):

```sh
curl -fsSL https://raw.githubusercontent.com/ethanhamilthon/jin/main/install.sh | sh
```

Environment options: `JIN_VERSION=v0.5` pins a release, `JIN_INSTALL_DIR=/some/dir`
chooses the target directory. Release builds keep their data in `~/.jin`.

## Benchmark

A full benchmark is still to come. These are early results from a small run: Aider Polyglot,
10 Python tasks, one run for each agent and task (50 runs in all), all agents on the same
model (`gpt-6-luna`, effort `high`). Numbers are the average of one run.

| Agent | Pass | Agent time, s | Requests | Input | Cached | Output | Reasoning |
| --- | --- | --- | --- | --- | --- | --- | --- |
| pi | 0.60 | 60 | 6.3 | 15 706 | 9 370 | 1 593 | 770 |
| opencode | 0.60 | 101 | 11.0 | 72 617 | 57 293 | 2 334 | 683 |
| codex | 0.50 | 48 | 4.1 | 43 592 | 34 842 | 1 645 | 835 |
| omp | 0.50 | 97 | 11.2 | 89 729 | 76 595 | 2 625 | 1 068 |
| jin (v0.4) | 0.50 | 56 | 6.1 | 16 929 | 9 984 | 1 708 | 816 |

Pass is the share of tasks with reward 1.0. The gap between agents is one task in ten,
which is within noise, so read this as "comparable", not as a ranking. What stands out is
the cost: jin sent about 17k input tokens per run, against 44k to 90k for codex, opencode
and omp. Tokens come from the `usage` field of the API responses, taken through a logging
proxy, not from the agents' own logs. Reasoning is part of Output. Agent time is the run
phase only, without building the image or installing the agent.

Caveats: ten tasks and one run each is a small sample; 11 of the 50 runs did not reach the
model (install timeout, network) and were restarted, so their times may be a little low; the
jin build was a local v0.4 build, not a release. Setup, task list and raw data are in the
`tbench` project.

## Release

Pushing a tag builds and publishes the release (`.github/workflows/release.yml`):

```sh
git tag v0.5 && git push origin v0.5
```

The workflow runs `make check`, then `scripts/build-release.sh <tag>`, which writes
`dist/jin_<os>_<arch>.tar.gz` and `dist/checksums.txt`, and attaches them to a GitHub
release. `make release VERSION=v0.5` builds the same archives locally.

## Build

Requires Go 1.27 or newer.

```sh
make build        # bin/jin, uses ~/.jin-dev for data
make build-prod   # bin/jin, uses ~/.jin for data
make install      # build-prod and copy to /usr/local/bin
make check        # go test + go vet
```

Run `jin` from the project directory you want to work on.

jin has no web search. If you need one, write a CLI and describe it in the global
`AGENTS.md` (see Context below): the model runs it through `bash`.

## Headless

Only a bare `jin` opens the TUI; every other command runs without it, and an unknown one
says so and points to `jin --help`. `jin -p "prompt"` runs one request and prints the
answer, for scripts and CI. See [docs/headless.md](docs/headless.md).

```
jin -p "summarize the last commit"
git diff | jin -p --format json "review this" | tail -1 | jq -r .result
jin refresh-models --efforts && jin models
```

- `jin models` and `jin refresh-models [--efforts]` list the models of your provider.
- `JIN_BASE_URL`, `JIN_API_KEY`, `JIN_PROVIDER_KIND` (`openai` or `anthropic`), `JIN_MODEL` and
  `JIN_EFFORT` override the saved settings
  for headless runs only. The TUI ignores them and jin never saves them.
- `JIN_DEPTH` counts nested `jin -p` runs; at depth 3 `jin -p` exits with an error.

## System prompts and hooks

`#plan`, `#review` and `#subagents` are built into jin. They are marked `system`: you can
switch them off, not edit or delete them. Any prompt can be switched off with `t` in
`/prompts`. Prompts you add are files in `~/.jin/prompts`.

Prompts, hooks and the system prompt can run shell commands: `{{git status --short}}` is
replaced by the output when a session starts. `/system-prompt` edits the system prompt, and
the prompts for compaction and handoff, in `~/.jin/system-prompt.md`. See
[docs/prompts-and-hooks.md](docs/prompts-and-hooks.md).

## Async tasks

The agent can start a command in the background with `jin async run "<cmd>" --session <id>`
and keep working. A daemon runs the task, also after you close jin. When it ends, the
result comes back to the agent as a message, even if the agent has already finished its turn.
Sub-agents use the same call to wake their parent. `/async-tasks` shows what runs. See
[docs/async.md](docs/async.md).

## First run

1. Type `/provider`, press `a`, choose the kind (OpenAI-compatible or Anthropic-compatible),
   give it a name, the base URL (for example `https://api.openai.com/v1`, or
   `https://api.anthropic.com` for Anthropic) and the API key.
2. Pick a model and a reasoning effort.
3. Type a message.

`Esc` only closes what is open (a list, a panel, `/bash`); with nothing open it does nothing.
To stop a running request press `Ctrl+C` or type `/stop`.

## Slash commands

Type `/` anywhere in the input: a list opens with an icon and a short description for
each command. `Tab` or `Enter` runs the highlighted one and cuts `/name` out of the draft;
the rest of the draft stays. The `/` must start the text or follow whitespace, so paths
like `/usr/bin` are plain text. `Esc` closes the list and keeps the text.

| Group | Commands |
| --- | --- |
| Menus | `/sessions`, `/prompts`, `/hooks` (one panel, `←`/`→` switch tabs) |
| Settings | `/model`, `/scope`, `/provider`, `/sound`, `/tools`, `/change-editor` |
| Background | `/async-tasks` |
| Actions | `/compact`, `/handoff`, `/stop`, `/new` (the draft moves into the new session), `/quit` |
| Input | `/clear`, `/copy`, `/edit`, `/todo` |
| Modes | `/tui <command>` runs a full-screen program (`/tui lazygit`); `/bash` opens a shell line until `Esc` |

`/bash` runs commands in the working directory and shows the output in the chat only; the
model never sees it. In a panel, `↑`/`↓` pick a row and `←`/`→` change a value (Sound) or
switch tabs.

`/scope` turns models of the provider on and off (`Enter` toggles). Only enabled models
show up in `/model` and in the `Ctrl+M` rotation. `Ctrl+M`
needs a terminal that tells it apart from `Enter` (kitty keyboard protocol:
kitty, Ghostty, WezTerm, foot, recent iTerm2 and Alacritty).

## Files

Type `@` to attach a file: `@notes.md`, `@src/main.go`, `@~/docs/plan.md`. The list stays
open while the text after `@` is a valid path, and `Tab` on a directory goes inside it. Put a
name with spaces in quotes: `@"my file.md"`. On send, every `@path` that is a real file is
replaced by its full path and listed at the end of the request:

```
<attached-files>
<file path="/Users/me/project/notes.md"/>
</attached-files>
```

The model reads the files itself. A path that does not exist stays plain text.

## Providers

`/provider` lists the saved providers (`a` add, `d` delete, `Enter` use). A session keeps the
provider it started with. Two kinds: OpenAI-compatible (`/chat/completions`) and
Anthropic-compatible (`/v1/messages`). Reasoning levels come from the provider when it tells
them; otherwise `low`, `medium` and `high` are offered. For Anthropic a level is sent as
adaptive thinking with that effort. v0.2 settings are migrated on first start: the old
provider becomes the provider `default`. See [docs/how-it-works.md](docs/how-it-works.md).

## Context

The system prompt includes every `AGENTS.md` that applies: the global one
(`~/.jin/AGENTS.md`, `~/.jin-dev/AGENTS.md` for source builds), the ones in
parent directories (marked as not part of the current project) and the one in
the project directory. The start screen lists the files that were used.

Jin does not edit these files; open them in your own editor.

`/compact` asks the model to summarize the session; the chat shows
a divider and the model continues from the summary. The full history stays
saved. It also runs on its own when the context reaches 80% of the model's
window (known from the OpenRouter and LiteLLM catalogues). `/handoff` has the model
write a brief and opens a new session with it in the input, ready to edit.

The notification on a final answer is the macOS system sound (`afplay`) with the
volume from `/sound`; other systems get the terminal bell, which has no
volume. Sound has three rows: Toggle (on, off), When (always, on blur) and Volume
(10% to 100%). "On blur" needs a terminal that reports focus changes.

## Keys

Input

| Key | Action |
| --- | --- |
| `Enter` | send |
| `Shift+Enter` | new line |
| `↑` / `↓` | move the cursor between lines (the input scrolls) |
| `Ctrl+V` | paste at the cursor |
| `Ctrl+C` | copy the selection, otherwise interrupt the running request (or `/bash` command) |
| `Ctrl+M` | next model from the scope |
| `Ctrl+O` | next folding mode |
| `Ctrl+T` | edit the todo list in the editor |
| `/` `#` `@` | start a command, a prompt name or a file path |
| `Esc` | close an open list or panel; does nothing otherwise |

Shortcuts work in any keyboard layout: `Ctrl+м` on a Russian layout is `Ctrl+V`.

Panel

| Key | Action |
| --- | --- |
| `Esc` | close the panel |
| `←` / `→` | switch tabs |
| `↑` / `↓` | move between items, wrapping around |
| `Enter` | select |

In a list without action keys the search is always focused: type to filter.
Lists with action keys (Prompts, Hooks, Providers) read plain letters as actions, and `/` opens
search. `Esc` closes the whole panel, including an active search.
Scroll the chat with the mouse wheel.

Select text with the mouse to copy it. Click a link in an answer to open it.

`ask_user` questions: `↑`/`↓` choose, `Enter` answers, `←`/`→` go back and forward between
questions to change an answer, `Space` ticks an option when the question allows several.

## Folding

`Ctrl+O` steps through four modes in a loop, in every session. The mode is saved
in the settings and kept between runs.

1. Tool output: everything, plus the last 5 lines of each `bash` output and the
   changed lines of each `edit` and `write` (red removed, green added).
2. Everything: messages, reasoning and tool calls.
3. No tool calls: messages and reasoning.
4. Messages only: just your messages and the agent's replies.

The last line of the chat says what the next `Ctrl+O` does. While the agent works
and something is hidden, that line shows `⠋ working...` with the last thing it did
(a tool call, or in mode 4 its reasoning) before the hint.

## Prompts

Prompts are markdown files in `~/.jin/prompts` (`~/.jin-dev/prompts` for source
builds). A folder is part of the name: `review/security.md` is `#review/security`.

- Type `/prompts`. `Enter` edits, `a` adds, `d` deletes,
  `e` changes the editor (nano, vim or hx, asked on first use), `/` searches.
- Type `#` in the input to autocomplete a name. `Tab` or `Enter` completes it.
- On send, every `#name` that matches a prompt is added to the request inside
  `<pasted-prompts>`, as `<prompt name="...">`. The chat shows only what you typed.

## Hooks

A hook is a prompt that goes into the system prompt when a session starts. Hooks are
global markdown files in `~/.jin/hooks` (`~/.jin-dev/hooks` for source builds).

- Type `/hooks`. `Enter` edits, `a` adds, `d` deletes,
  `t` switches a hook on or off (a new hook is on), `e` changes the editor, `/` searches.
- Enabled hooks are added in alphabetical order, as plain text, before the `AGENTS.md`
  block. Empty hooks add nothing.
- Edits apply to new sessions only. The start screen lists the enabled hooks after Context.

## System prompt

The system prompt is built when a session starts. Its text is the built-in default or the
`# system` section of `~/.jin/system-prompt.md` (type `/system-prompt` to edit it; the same
file holds the `# compact` and `# handoff` prompts). Jin adds the tool list, your hooks, a
pointer to the docs, the `jin async` instructions (with the `bash` tool and a session id) and
the `AGENTS.md` files. Details: [docs/how-it-works.md](docs/how-it-works.md).

## Data

Everything lives in `~/.jin` (or `~/.jin-dev` for source builds): the SQLite
database `jin.db` with sessions and settings, and pasted images. The provider API key is stored in the database as plain text.
The files are created with `0600` permissions. Several jin processes can use the
same directory at once: the database is in WAL mode and writers wait for each
other.

## Layout

```
main.go               wiring
internal/core         agent loop, system prompt
internal/provider     OpenAI-compatible and Anthropic-compatible streaming client
internal/files        @file mentions: find, complete, resolve, XML block
internal/tools        read, write, edit, bash, ask_user, todo
internal/cli          which command a command line means
internal/headless     jin -p, jin models, jin refresh-models
internal/async        jin async, the background-task daemon
internal/upgrade      one-time steps when a new version starts
internal/dyn          {{commands}} in prompts
internal/sysprompt    ~/.jin/system-prompt.md
internal/startup      builds a session's prompts before its agent starts
internal/store        SQLite sessions and settings
internal/ui           terminal interface
```

Documentation for users and agents is in [docs/](docs/README.md). Jin tells the agent where to
read it when you ask about jin. Development rules are in [AGENTS.md](AGENTS.md).

## Safety

Tools run without confirmation. `bash`, `write` and `edit` act directly on your
machine, so run jin only in projects where that is acceptable.
