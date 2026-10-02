# jin

A minimal TUI coding agent written in GO.

- Any OpenAI-compatible API
- Input-first TUI with an Esc-toggle tabbed panel
- Streaming / Markdown rendering
- Tools: `read`, `write`, `edit`, `bash`, `ask_user`, `todo` (each can be switched off in Settings → Tools)
- Images: `read` shows a picture to the model (png, jpeg, gif, webp, bmp). Pasting an image
  saves it to disk and types its path, so ask the agent to read that path
- Several sessions at once, saved per directory in SQLite
- Reusable prompts: type `#name` in the input, manage them from the `Esc` menu
- Several jin processes can run at once and share one database

## Install

macOS and Linux, arm64 and amd64. The script downloads the latest release, checks its
SHA-256 and installs `jin` into `/usr/local/bin` (or `~/.local/bin` when that is not
writable):

```sh
curl -fsSL https://raw.githubusercontent.com/ethanhamilthon/jin/main/install.sh | sh
```

Environment options: `JIN_VERSION=v0.1` pins a release, `JIN_INSTALL_DIR=/some/dir`
chooses the target directory. Release builds keep their data in `~/.jin`.

## Release

Pushing a tag builds and publishes the release (`.github/workflows/release.yml`):

```sh
git tag v0.1 && git push origin v0.1
```

The workflow runs `make check`, then `scripts/build-release.sh <tag>`, which writes
`dist/jin_<os>_<arch>.tar.gz` and `dist/checksums.txt`, and attaches them to a GitHub
release. `make release VERSION=v0.1` builds the same archives locally.

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

## First run

1. Press `Esc`, open Settings, choose Provider, and enter the base URL and API key of an
   OpenAI-compatible provider (for example `https://api.openai.com/v1`).
2. Pick a model and a reasoning effort.
3. Press `Esc` to close any open panel and type a message.

Press `Esc` to switch between the input and the panel. Its tabs:

- Commands: new session, interrupt, compact, handoff, quit
- Sessions: sessions of this directory (green dot: the open one, blinking blue
  dot: answering, blue dot: an unread answer)
- Prompts: reusable prompts (see Prompts below)
- Context: the `AGENTS.md` files of the system prompt (see Context below)
- Settings: select model, scope models, provider, sound, tools, jin docs, editor

- Input: focus input, clear, copy, paste, edit in editor

Choosing an action returns to the input unless it opens another panel.
Input actions preserve the draft except Clear; Paste inserts at the cursor.
Every panel above the input is 6 rows high and scrolls. In Sound, `↑`/`↓`
pick a row and `←`/`→` change its value.

Scope models turns models of the provider on and off (`Enter` toggles). Only
enabled models show up in Select model and in the `Ctrl+M` rotation. `Ctrl+M`
needs a terminal that tells it apart from `Enter` (kitty keyboard protocol:
kitty, Ghostty, WezTerm, foot, recent iTerm2 and Alacritty).

## Context

The system prompt includes every `AGENTS.md` that applies: the global one
(`~/.jin/AGENTS.md`, `~/.jin-dev/AGENTS.md` for source builds), the ones in
parent directories (marked as not part of the current project) and the one in
the project directory. The start screen lists the files that were used.

The Context tab of the `Esc` menu lists every one of these files, empty ones too.
`Enter` edits a file. Files cannot be deleted. `a` creates an `AGENTS.md` in the
current directory, offered only when there is none.

Compact (Commands tab) asks the model to summarize the session; the chat shows
a divider and the model continues from the summary. The full history stays
saved. It also runs on its own when the context reaches 80% of the model's
window (known from the OpenRouter and LiteLLM catalogues). Handoff has the model
write a brief and opens a new session with it in the input, ready to edit.

The notification on a final answer is the macOS system sound (`afplay`) with the
volume from Settings → Sound; other systems get the terminal bell, which has no
volume. Sound has three rows: Toggle (on, off), When (always, on blur) and Volume
(10% to 100%). "On blur" needs a terminal that reports focus changes.

## Keys

Input

| Key | Action |
| --- | --- |
| `Esc` | open the panel |
| `Enter` | send |
| `Shift+Enter` | new line |
| `↑` / `↓` | move the cursor between lines (the input scrolls) |
| `Ctrl+V` | paste at the cursor |
| `Ctrl+C` | interrupt the running request |
| `Ctrl+M` | next model from the scope |
| `Ctrl+O` | next folding mode |
| `Ctrl+T` | edit the todo list in the editor |

Panel

| Key | Action |
| --- | --- |
| `Esc` | close any panel and focus input |
| `←` / `→` | switch tabs |
| `↑` / `↓` | move between items, wrapping around |
| `Enter` | select |

In a list without action keys the search is always focused: type to filter.
Lists with action keys (Prompts) read plain letters as actions, and `/` opens
search. `Esc` closes the whole panel, including an active search.
Scroll the chat with the mouse wheel.

Select text with the mouse to copy it.

## Folding

`Ctrl+O` steps through three modes in a loop, in every session. The mode is saved
in the settings and kept between runs.

1. Everything: messages, reasoning and tool calls.
2. No tool calls: messages and reasoning.
3. Messages only: just your messages and the agent's replies.

The last line of the chat says what the next `Ctrl+O` does. While the agent works
and something is hidden, that line shows `⠋ working...` with the last thing it did
(a tool call, or in mode 3 its reasoning) before the hint.

## Prompts

Prompts are markdown files in `~/.jin/prompts` (`~/.jin-dev/prompts` for source
builds). A folder is part of the name: `review/security.md` is `#review/security`.

- Open the Prompts tab of the `Esc` menu. `Enter` edits, `a` adds, `d` deletes,
  `e` changes the editor (nano, vim or hx, asked on first use), `/` searches.
- Type `#` in the input to autocomplete a name. `Tab` or `Enter` completes it.
- On send, every `#name` that matches a prompt is added to the request inside
  `<pasted-prompts>`, as `<prompt name="...">`. The chat shows only what you typed.

## Hooks

A hook is a prompt that goes into the system prompt when a session starts. Hooks are
global markdown files in `~/.jin/hooks` (`~/.jin-dev/hooks` for source builds).

- Open Context → Hooks in the `Esc` menu. `Enter` edits, `a` adds, `d` deletes,
  `t` switches a hook on or off (a new hook is on), `e` changes the editor, `/` searches.
- Enabled hooks are added in alphabetical order as plain text, before the `AGENTS.md`
  block. Empty hooks add nothing.
- Edits apply to new sessions only. The start screen lists the enabled hooks after Context.

## System prompt

The system prompt is `internal/core/system_prompt.md`. It is embedded at build
time, so rebuild after editing. Available placeholders:

- `{{dir}}` working directory
- `{{os}}` operating system and architecture
- `{{date}}` current date
- `{{hooks}}` enabled hooks, followed by a blank line, or nothing when there are none

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
internal/provider     OpenAI-compatible streaming client
internal/tools        read, write, edit, bash
internal/store        SQLite sessions and settings
internal/ui           terminal interface
```

Documentation for users and agents is in [docs/](docs/README.md). Enable Settings → Jin docs
to let the agent read it when you ask about jin. Development rules are in [AGENTS.md](AGENTS.md).

## Safety

Tools run without confirmation. `bash`, `write` and `edit` act directly on your
machine, so run jin only in projects where that is acceptable.
