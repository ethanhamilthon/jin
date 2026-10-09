# Jin docs

Jin is a minimal terminal coding agent written in Go. These docs describe how it works
and how to bend it to your needs. Read only the file that matches your question.

| File | Covers |
| --- | --- |
| [about.md](about.md) | What jin is, its multi-project workspace and its philosophy |
| [how-it-works.md](how-it-works.md) | Agent loop, providers, retries, tools, project-aware SQLite state, system prompt, context, compact and handoff |
| [headless.md](headless.md) | `jin -p`, sessions, flags, JSON output, exit codes, environment, `jin models` and `jin refresh-models` |
| [tasks.md](tasks.md) | Background tasks: the `task` tool, results, the note before the final answer and `/tasks` |
| [tui.md](tui.md) | Projects, sessions, prompts and hooks panels, split panes, work-view commands, `$` shell input, `/tui`, keys, mouse, folding and `todo` |
| [web.md](web.md) | `jin web`: the browser UI, its flags, access token, panes, composer, settings and how to build it |
| [extending.md](extending.md) | CLI hooks and prompts, a validated `searchctl` example, sharing hooks with a team |
| [prompts-and-hooks.md](prompts-and-hooks.md) | Reusable `#prompts`, built-ins, `{{commands}}`, the System prompt and Hooks rows of `/settings`, `AGENTS.md`, CLI integrations |
| [gallery.md](gallery.md) | Hooks and prompts for git, tests, release checks, skill-style prompts, sub-agents, `gh`, and a clearly labeled MCP adapter template |
| [benchmarks.md](benchmarks.md) | Latest 20-task benchmark (jin v0.7.2 and v0.7.0, codex, pi, opencode), earlier 10-task run, method and caveats |
| [database.md](database.md) | SQLite project and session schema, settings and `sqlite3` recipes |

Source code: https://github.com/ethanhamilthon/jin
