# jin

A minimal coding agent for the terminal, written in Go. It talks to any
OpenAI-compatible API and works on the project in your current directory.

- Vim-style TUI with NORMAL and INSERT modes
- Streaming replies with markdown rendering and reasoning effort selection
- Tools: `read`, `write`, `edit`, `bash`, `websearch`
- Several sessions at once, saved per directory in SQLite
- Reusable prompts: type `#name` in the input, manage them with `r`
- Web search through DuckDuckGo (no key), Brave or Tavily

## Build

Requires Go 1.27 or newer.

```sh
make build        # bin/jin, uses ~/.jin-dev for data
make build-prod   # bin/jin, uses ~/.jin for data
make install      # build-prod and copy to /usr/local/bin
make check        # go test + go vet
```

Run `jin` from the project directory you want to work on.

## First run

1. Press `p` in NORMAL mode and enter the base URL and API key of an
   OpenAI-compatible provider (for example `https://api.openai.com/v1`).
2. Pick a model and a reasoning effort.
3. Press `i` and type a message.

Press `Space` in NORMAL mode to see all commands instead of remembering keys.

## Keys

INSERT

| Key | Action |
| --- | --- |
| `Esc` | back to NORMAL |
| `Enter` | send |
| `Shift+Enter` | new line |
| `Ctrl+C` | interrupt the running request |

NORMAL

| Key | Action |
| --- | --- |
| `Space` | open the command menu |
| `i` / `a` | insert at start / end of the draft |
| `j` / `k`, arrows | scroll |
| `Ctrl+D` / `Ctrl+U` | scroll half a page |
| `g` / `G` | jump to top / bottom |
| `n` | new session |
| `s` | sessions of this directory |
| `m` | model and reasoning effort |
| `p` | provider |
| `r` | prompts: add, edit, delete |
| `w` | web search backend |
| `x` | interrupt |
| `q` | quit |

In any list: `j`/`k` move, `/` filters, `Enter` selects, `Esc` closes.

Select text with the mouse to copy it.

## Prompts

Prompts are markdown files in `~/.jin/prompts` (`~/.jin-dev/prompts` for source
builds). A folder is part of the name: `review/security.md` is `#review/security`.

- Press `r` in NORMAL mode to list prompts. `Enter` edits, `a` adds, `d` deletes,
  `e` changes the editor (nano, vim or hx, asked on first use).
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
database `jin.db` with sessions and settings, and pasted images. Provider and
search API keys are stored in the database as plain text. The files are created
with `0600` permissions.

## Layout

```
main.go               wiring
internal/core         agent loop, system prompt
internal/provider     OpenAI-compatible streaming client
internal/tools        read, write, edit, bash, websearch
internal/search       search backends
internal/store        SQLite sessions and settings
internal/ui           terminal interface
```

Development rules are in [AGENTS.md](AGENTS.md).

## Safety

Tools run without confirmation. `bash`, `write` and `edit` act directly on your
machine, so run jin only in projects where that is acceptable.
