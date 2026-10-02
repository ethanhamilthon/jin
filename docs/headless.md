# Headless mode

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
| `--tools a,b` | keep only these tools |
| `--exclude-tools a,b` | remove these tools |
| `--no-tools` | no tools at all |
| `--timeout <d>` | stop after a duration (`90s`, `10m`); a bare number is seconds. No default |

`-c` and `--session` cannot be combined.

## Input

The prompt is the positional text. When stdin is a pipe it is appended in a
`<stdin>...</stdin>` block (10 MB cap). Without prompt text the whole prompt is read
from stdin. `#name` prompts are not expanded in headless mode.

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
| `JIN_PROVIDER_KIND` | `openai` or `anthropic`; overrides the kind of the saved provider |
| `JIN_MODEL`, `JIN_EFFORT` | model and effort; used by `jin -p` only |

`jin models` and `jin refresh-models` use only the provider variables.

Order: flag, environment, session record, saved settings. `JIN_DEPTH` counts nested
headless runs; at depth 3 `jin -p` exits with `subagent depth limit`.

## Exit codes

`0` success, `1` any error (usage, config, provider, timeout), `130` interrupted.

## Subagents

The `#subagents` prompt teaches the agent to start `jin -p --no-session --model <id>`
in the background through `bash`, one per independent task, and to track them in `todo`.
The launch call returns at once. The agent does its own work, then checks every 10-20
seconds for the `N.exit` file that each sub-agent writes when it ends. Output goes to
files, because `bash` output is cut at 16 KB. The point is speed: when parallel agents
would not make the task faster, the prompt tells the agent to say so and work alone.
Background agents outlive the `bash` call that started them; jin stops the leftovers
when it exits. The models for sub-agents are set in the `Models` fields at the top of
`~/.jin/prompts/subagents.md` (see [prompts-and-hooks.md](prompts-and-hooks.md)). Headless runs never expand `#name`, so a sub-agent cannot start
sub-agents this way; `JIN_DEPTH` caps nesting at 3.

## Models

- `jin refresh-models` fetches the model list and caches it (setting `models.cache`).
  `--efforts` also probes the reasoning levels of every model (`models.levels`).
- `jin models` prints the cached list, fetching it once if the cache is empty: `id`, or
  `id<TAB>low,medium,high` when levels are known. `--format json` prints
  `[{"id":"...","efforts":[...]}]`; `efforts` is left out when the levels are unknown.
