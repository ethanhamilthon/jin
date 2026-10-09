# Parity: TUI, jin web and `jin -p`

The agent is the same in all three modes: `internal/agentkit` builds it, and every mode renders
the system prompt, hooks and `#prompts` with `internal/startup`. This page lists what differs
and why. Interface differences (themes, panes, key chords) are left out. `gap` means a
feature that should exist and does not yet; `by design` means the mode cannot or should not
have it.

## Agent

| Feature | TUI | Web | Headless | Status |
| --- | --- | --- | --- | --- |
| Tools: `read`, `write`, `edit`, `bash`, `task`, `tell_user`, `todo` | yes | yes | yes | same |
| `ask_user` | yes | yes | no | by design: nobody can answer |
| `bash` past its timeout | moves to a background task | moves to a background task | killed | by design: the process ends with the run |
| Background tasks | outlive a run | outlive a run | end with the run | by design |
| Working directory of the agent | session path | session path | `--cwd` or the current directory | same |
| Hooks, `AGENTS.md`, nested `AGENTS.md` | yes | yes | yes | same |
| `#prompt` | yes | yes | yes | same |
| System prompt refresh after compaction or a cold cache | yes | yes | yes | closed in v0.10: headless had no refresher |
| Tool set narrowing | `/settings` | Settings | `--tools`, `--exclude-tools`, `--no-tools` | same, a flag only removes tools |
| Budgets (`--max-cost`, `--max-turns`) | no | no | yes | by design: unattended runs |
| Pictures | paste with `Ctrl+V` | paste, drop, pick | the `read` tool reads a picture | by design: headless input is text |
| Streaming of deltas | yes | yes | no | by design: one result |

## Session commands

| Command | TUI | Web | Headless |
| --- | --- | --- | --- |
| Continue a session | `/sessions` | sidebar | `-c`, `--session` |
| Search sessions | `/sessions` | sidebar search | `jin sessions search` |
| Export | no | Export as Markdown | `jin export` |
| Context report | `/context` | Context | gap: `jin session context` (v0.10) |
| Compact | `/compact` | Compact | gap: `jin session compact` (v0.10) |
| Handoff | `/handoff` | Handoff | gap: `jin session handoff` (v0.10) |
| Rewind | `/rewind` | Rewind | gap: `jin session rewind` (v0.10) |
| Undo | `/undo` | Undo | gap: `jin session undo` (v0.10) |
| Reload prompts | `/reload` | Reload prompts | gap: `jin session reload` (v0.10) |
| Stop | `/stop` | Stop | `Ctrl+C`, `--timeout` |
| Background tasks list | `/tasks` | Tasks | by design: tasks end with the run |

## Settings

| Setting | TUI | Web | Headless |
| --- | --- | --- | --- |
| Providers | `/provider` | Providers | `--provider`, `JIN_BASE_URL`, `JIN_API_KEY` |
| Model and effort | `/model` | model button | `--model`, `--effort`, `JIN_MODEL` |
| Tools, prompts, hooks, system prompt | `/settings` | Settings | read from the same data folder |
| Hooks trust of a project | asked when a hook is enabled in the Hooks panel | asked when the project opens | project hooks run only when trusted before |
| Theme, accent, sound, motion | yes | accent and sound | by design |
| Data folder | `/settings` | Settings | by design |

## Open points

- Export has no TUI command. Add it, or mark it by design, in the brainstorm of v0.11.
- Headless cannot change settings. Intentional: the TUI and web are the place for them.
