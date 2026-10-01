# jin

A minimal TUI coding agent written in GO.

- Any OpenAI-compatible API
- Vim-style TUI with NORMAL and INSERT modes
- Streaming / Markdown rendering
- Tools: `read`, `write`, `edit`, `bash`
- Several sessions at once, saved per directory in SQLite
- Reusable prompts: type `#name` in the input, manage them from the `Space` menu
- Several jin processes can run at once and share one database

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

1. Press `Space`, open Settings, choose Provider, and enter the base URL and API key of an
   OpenAI-compatible provider (for example `https://api.openai.com/v1`).
2. Pick a model and a reasoning effort.
3. Press `i` and type a message.

NORMAL mode has no command hotkeys: everything goes through `Space`. Its tabs:

- Commands: new session, interrupt, compact, handoff, quit
- Sessions: sessions of this directory (green dot: the open one, blinking blue
  dot: answering, blue dot: an unread answer)
- Settings: select model, scope models, provider, prompts, sound, editor

Menus stay in NORMAL mode after a choice; only New session, a picked session and
a handoff brief switch to INSERT. Every panel above the input is 6 rows high and
scrolls. In Sound, `↑`/`↓` pick a row and `←`/`→` change its value.

Scope models turns models of the provider on and off (`Enter` toggles). Only
enabled models show up in Select model and in the `Ctrl+M` rotation. `Ctrl+M`
needs a terminal that tells it apart from `Enter` (kitty keyboard protocol:
kitty, Ghostty, WezTerm, foot, recent iTerm2 and Alacritty).

## Context

The system prompt includes every `AGENTS.md` that applies: the global one
(`~/.jin/AGENTS.md`, `~/.jin-dev/AGENTS.md` for source builds), the ones in
parent directories (marked as not part of the current project) and the one in
the project directory. The start screen lists the files that were used.

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

INSERT

| Key | Action |
| --- | --- |
| `Esc` | back to NORMAL |
| `Enter` | send |
| `Shift+Enter` | new line |
| `↑` / `↓` | move the cursor between lines (the input scrolls) |
| `Ctrl+C` | interrupt the running request |

NORMAL

| Key | Action |
| --- | --- |
| `Space` | open the menu; `←`/`→` switch tabs |
| `Ctrl+M` | next model from the scope (INSERT and NORMAL) |
| `i` / `a` | insert at start / end of the draft |
| `j` / `k`, arrows | scroll |
| `Ctrl+D` / `Ctrl+U` | scroll half a page |
| `g` / `G` | jump to top / bottom |

In any list the search is always focused: just type to filter, `↑`/`↓` move
(wrapping around, so `↑` on the first item jumps to the last), `Enter` selects,
`Esc` closes.

Select text with the mouse to copy it.

## Prompts

Prompts are markdown files in `~/.jin/prompts` (`~/.jin-dev/prompts` for source
builds). A folder is part of the name: `review/security.md` is `#review/security`.

- Open Prompts from Settings in the `Space` menu. `Enter` edits, `Ctrl+A` adds, `Ctrl+D` deletes,
  `Ctrl+E` changes the editor (nano, vim or hx, asked on first use).
- Type `#` in the input to autocomplete a name. `Tab` or `Enter` completes it.
- On send, every `#name` that matches a prompt is added to the request inside
  `<pasted-prompts>`, as `<prompt name="...">`. The chat shows only what you typed.

## System prompt

The system prompt is `internal/core/system_prompt.md`. It is embedded at build
time, so rebuild after editing. Available placeholders:

- `{{dir}}` working directory
- `{{os}}` operating system and architecture
- `{{date}}` current date

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

Development rules are in [AGENTS.md](AGENTS.md).

## Safety

Tools run without confirmation. `bash`, `write` and `edit` act directly on your
machine, so run jin only in projects where that is acceptable.
