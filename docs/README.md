# Jin docs

Jin is a minimal terminal coding agent written in Go. These docs describe how it works
and how to bend it to your needs. Read only the file that matches your question.

| File | Covers |
| --- | --- |
| [about.md](about.md) | What jin is and its philosophy |
| [how-it-works.md](how-it-works.md) | Agent loop, tools (`read`, `write`, `edit`, `bash`, `ask_user`, `todo`) and `/tools`, system prompt (with the session id), context, compact, handoff, data directory |
| [headless.md](headless.md) | Which command does what (`jin` alone opens the TUI), `jin -p` for scripts and sub-agents: flags, JSON output, exit codes, env vars, `jin models`, `jin refresh-models` |
| [async.md](async.md) | Background tasks for the agent: `jin async run/check/input/stop`, the daemon, results that wake the agent, `/async-tasks` |
| [tui.md](tui.md) | Moving around the TUI: input, slash commands, `@file` mentions, `/bash`, `/tui`, keys (`Ctrl+T` edits the todo list), mouse, folding, `ask_user` and `todo` blocks |
| [prompts-and-hooks.md](prompts-and-hooks.md) | Reusable `#prompts`, the built-in `#plan`, `#review`, `#subagents`, `{{commands}}` in prompts, `/system-prompt`, hooks, `AGENTS.md`, adding your own tools as CLIs |
| [database.md](database.md) | SQLite schema and `sqlite3` recipes for sessions, usage and settings |

Source code: https://github.com/ethanhamilthon/jin
