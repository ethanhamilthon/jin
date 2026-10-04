# Headless mode

## Commands

Only a bare `jin` opens the TUI. Every other command runs without it:

| Command | Does |
| --- | --- |
| `jin -p ...` | one request, see below |
| `jin models`, `jin refresh-models` | model list, see below |
| `jin export <id> [--md\|--json]` | print a saved session, see below |
| `jin sessions list\|search` | list or search saved sessions, see below |
| `jin hooks add\|list` | share hooks, see [prompts-and-hooks.md](prompts-and-hooks.md) |
| `jin async run\|check\|input\|stop` | background tasks, see [async.md](async.md) |
| `jin update [--check]` | install the latest release over the running binary (checks its SHA-256); `--check` only tells whether one exists |
| `jin --version`, `jin --help` | version and usage |

Anything else prints `jin: unknown command "x"` and `Run 'jin --help' for usage.` and exits
with `2`. `jin --help`, `--version` and unknown commands never touch the database.

`jin export <session-id>` prints a saved session to stdout: Markdown by default (user
and assistant messages, tool calls and results in folded blocks), or with `--json` the
session fields and the raw messages. A unique prefix of the id is enough. Use it for bug
reports, sharing, or feeding a session to another agent:
`jin export 3f2a --md > session.md`.

`jin sessions list [--all] [--format json]` lists saved sessions in the current directory
(or across all directories with `--all`).

`jin sessions search <words...> [--all] [--format json]` searches session titles and user
messages (case-insensitive LIKE, all words must match, at most 20 results, newest first).
Text output is tab separated: id, date, title, snippet.

`jin -p` runs one request without the TUI, prints the result and exits. Use it in
scripts, CI and from other agents.

```
jin -p [flags] [prompt...]
```

Flags work before and after the prompt. `--` ends the flags.

| Flag | Meaning |
| --- | --- |
| `--format text\|json` | output format, default `text` |
| `-c`, `--continue` | continue the latest session of the current directory |
| `--session <id>` | continue the session with this id |
| `--no-session` | save nothing; with `-c` or `--session` history is read only |
| `--model <id>`, `--effort <level>` | model and reasoning effort for this run |
| `--cwd <dir>` | run in this directory: the session is stored with that path, tools run there, `AGENTS.md` and project hooks come from there; `-c` continues the latest session of it |
| `--provider <id>` | use the saved provider with this id (see `/provider`); an unknown id is an error that lists the known ids |
| `--max-cost <usd>` | stop when this run has spent that many dollars (see Budgets) |
| `--max-turns <n>` | stop after `n` model requests (see Budgets) |
| `--tools a,b` | keep only these tools |
| `--exclude-tools a,b` | remove these tools |
| `--no-tools` | no tools at all |
| `--timeout <d>` | stop after a duration (`90s`, `10m`); a bare number is seconds. Total wall time: it also covers reading stdin and the commands of the system prompt. No default |

`-c` and `--session` cannot be combined. A session that another live jin process is running
(the TUI or another `jin -p`) is refused before anything is sent:
`session <id> is in use by process <pid>`, exit `1`. With `--no-session` it can still be read.

## Input

The prompt is the positional text. When stdin is a pipe it is appended in a
`<stdin>...</stdin>` block (10 MB cap). Without prompt text the whole prompt is read
from stdin. `#name` prompts are expanded like in the TUI (built-in and your own, except
disabled ones), and the `{{commands}}` in their bodies run like in the TUI. Bodies are
rendered only when the prompt text contains a `#`.

## Output

- `text`: stdout holds only the final answer. Tool calls and errors go to stderr, one
  line each (`bash: ls (120s)`).
- `json`: stdout is strict JSONL, logs go to stderr. Records: `session` (first),
  `message` (every message added to the history, as stored in the database) and
  `result` (last). There are no streaming deltas.

```
{"type":"result","is_error":false,"result":"...","session_id":"...","duration_ms":4210,
 "usage":{"input":1200,"output":80,"context":1280,"cost":0.0031}}
```

On failure `is_error` is `true` and an `error` string is added. Read the last line:

```
jin -p --format json "list the files" | tail -1 | jq -r .result
```

## Budgets

`--max-cost <usd>` and `--max-turns <n>` limit one run. A turn is one model request. Both count
from the start of this run, not from the totals of a continued session. After each usage update jin
compares; when a limit is reached it interrupts the agent, so the tool calls of that last request are
not waited for and the next request is never sent. The partial answer (the last assistant text) is
printed, the session is saved, the exit code is `3`, and with `--format json` the `result` record has
`"error":"budget reached: max-cost"` (or `max-turns`) and `is_error:true`. Also when the limit is
reached by the request that gives the final answer: the answer is printed, the exit code is still `3`. A
successful compaction counts as a request. `--max-cost` needs a price for the model
(the same catalogues as the cost in the status line); without one jin exits with `1` before sending
anything. A cost can pass the limit by the price of one request.

## Sessions and tools

Sessions are saved like in the TUI and show up in its session list. Tools: all enabled
tools (`/tools`) except `ask_user`, narrowed by `--tools`, `--exclude-tools` and
`--no-tools`; a flag cannot turn on a tool that is switched off. `todo` saves its list to
the database like in the TUI (not with `--no-session`); like every tool call it is
printed to stderr as `todo: ...`. There are no approvals.

## Environment

Read by headless commands only, never saved to the database:

| Variable | Meaning |
| --- | --- |
| `JIN_BASE_URL`, `JIN_API_KEY` | provider; each one overrides the saved value |
| `JIN_PROVIDER_KIND` | `openai`, `responses` or `anthropic`; overrides the kind of the saved provider |
| `JIN_MODEL`, `JIN_EFFORT` | model and effort; used by `jin -p` only |

`jin models` and `jin refresh-models` use only the provider variables. `jin models --provider <id>`
lists the models of that saved provider (the provider environment variables are ignored); for a provider that is not the active one the list is
fetched live and the model cache is left alone. `jin refresh-models` has no `--provider`.

Order for model and effort: flag, environment, session record, saved settings.

Order for the provider: `--provider`, environment (`JIN_BASE_URL`, `JIN_API_KEY`,
`JIN_PROVIDER_KIND`, each on its own), session record, saved active provider. With `--provider`
the provider variables are ignored: the flag names exactly that saved provider. With `--format json`
the first record (`session`) has `provider` (the id) and `endpoint` (the base URL without
credentials), so a script sees where the request went.

Provider: a run that continues a session (`-c`, `--session`) uses the provider recorded
in the session, not the active one. `JIN_BASE_URL`, `JIN_API_KEY` and `JIN_PROVIDER_KIND`
still override it. If the recorded provider was deleted, the run fails with
`session provider "<id>" no longer exists; set JIN_BASE_URL and JIN_API_KEY to override`
unless both variables are set. A new session records the active provider.

`JIN_DEPTH` counts nested
headless runs; at depth 3 `jin -p` exits with `subagent depth limit`.

## Exit codes

`0` success, `1` any error (usage, config, provider, timeout), `3` budget reached, `130` interrupted.

Only an error that ends the turn gives `1`. A smaller error, such as a failed
auto-compaction, is printed to stderr and does not change the code when the answer
arrives. If the history, file changes or usage cannot be saved, jin prints
`jin: could not save the session: <err>`, exits with `1` and, with `--format json`, adds
`"save_error"` to the `result` record while `is_error` stays `false`.

File changes of `write` and `edit` are saved like in the TUI, so `/undo` works on a
headless session.

## Subagents

The `#subagents` prompt teaches the agent to start `jin -p --no-session --model <id>`
through `jin async run`, one per independent task, and to track them in `todo`. The launch
returns at once and the agent does its own work; each sub-agent ends with
`jin async run "echo ..." --session <parent-id>`, which wakes the parent. See
[async.md](async.md). The point is speed: when parallel agents would not make the task
faster, the prompt tells the agent to say so and work alone. `JIN_DEPTH` caps nesting at 3, and the async daemon passes it on to the tasks.

## Models

- `jin refresh-models` fetches the model list and caches it (setting `models.cache`).
  `--efforts` also probes the reasoning levels of every model (`models.levels`).
- `jin models` prints only the models in your scope (`/scope`) for the active provider;
  an empty scope means every cached model. `--all` prints every cached model. The cache is
  fetched once if empty. Text, tab separated: `id`, input price and output price in dollars
  per 1M tokens, context window in tokens, efforts (`low,medium,high`). A value that is not
  known is empty. Prices come from the same catalogues as the cost in the status line; offline
  they are empty. `--format json` prints
  `[{"id":"...","input_per_mtok":1.25,"output_per_mtok":10,"context_window":400000,"efforts":[...]}]`;
  a field that is unknown is left out.
