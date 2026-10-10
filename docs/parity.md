# Parity: TUI, jin web and `jin -p`

The agent is the same in all three modes: `internal/agentkit` builds it, and every mode renders
the system prompt, hooks and `#prompts` with `internal/startup`. This page lists what differs
and why. Interface differences (themes, panes, key chords) are left out. `gap` means a
feature that should exist and does not yet; `by design` means the mode cannot or should not
have it.

## Agent

| Feature | TUI | Web | Headless | Status |
| --- | --- | --- | --- | --- |
| Tools: `read`, `grep`, `write`, `edit`, `bash`, `task` | yes | yes | yes | same |
| `ask_user` | yes | yes | no | by design: nobody can answer |
| `bash` past its timeout | moves to a background task | moves to a background task | killed | by design: the process ends with the run |
| Background tasks | outlive a run | outlive a run | end with the run | by design |
| Working directory of the agent | session path | session path | `--cwd` or the current directory | same |
| Hooks, `AGENTS.md` | yes | yes | yes | same |
| Exact text of every prompt part and tool schema | `/context full` | Context window | `jin sessions context <id> --full` | same |
| `#prompt` | yes | yes | yes | same |
| System prompt refresh after compaction or a cold cache | yes | yes | yes | closed in v0.10: headless had no refresher |
| Tool set narrowing | `/settings` | Settings | `--tools`, `--exclude-tools`, `--no-tools` | same, a flag only removes tools |
| Budgets (`--max-cost`, `--max-turns`) | no | no | yes | by design: unattended runs |
| Pictures | paste with `Ctrl+V` | paste, drop, pick | the `read` tool reads a picture | by design: headless input is text |
| Streaming of deltas | yes | yes | no | by design: one result |

## Session commands

| Command | TUI | Web | Headless |
| --- | --- | --- | --- |
| Continue a session | `/sessions`, `/projects` (picker: project, then sessions grouped by recency, all ages) | session picker (`switch` icon), same steps and groups | `-c`, `--session`, `jin sessions list --all` |
| Search sessions | in the picker, titles | session picker, titles | `jin sessions search` |
| Export | `/export` (writes a file in the project) | Export as Markdown | `jin export` |
| Session title | automatic, `/title` | automatic, Generate title | by design: keeps the first prompt |
| Context report | `/context` | Context | `jin sessions context` |
| Compact | `/compact` | Compact | `jin sessions compact` |
| Handoff | `/handoff` | Handoff | `jin sessions handoff` |
| Rewind | `/rewind` | Rewind | `jin sessions rewind` |
| Reload prompts | `/reload` | Reload prompts | `jin sessions reload` |
| Stop | `/stop` | Stop | `Ctrl+C`, `--timeout` |
| Background tasks list | `/tasks` | Tasks | by design: tasks end with the run |

## Settings

| Setting | TUI | Web | Headless |
| --- | --- | --- | --- |
| Providers (add, switch, delete, on/off) | `/provider` | Settings, Providers | `--provider`, `JIN_BASE_URL`, `JIN_API_KEY` |
| Enabled provider/model catalog | `/model`, Ctrl+M | model button | `jin models`; `--provider` selects one |
| Managed CLIProxyAPI connections | local Settings | local Settings; phones manage existing connections | uses already connected profiles |
| Proxy installation/version changes | Settings, CLIProxyAPI | Settings, CLIProxyAPI | settings are changed in TUI/web, by design |
| Model and effort | `/model` | model button | `--model`, `--effort`, `JIN_MODEL` |
| Tools, prompts, hooks, system prompt | `/settings` | Settings | read from the same data folder |
| Session titles (model, effort, prompt, after, rename at message 4) | `/settings` | Settings | not used |
| Projects: add, rename, archive, restore (no remove) | picker keys `a` `r` `x`, Settings, Archived projects | picker, Project pane, Settings, Archived projects | `jin projects` |
| No enabled provider | message on send | banner above the composer | error from the run |
| Hooks trust of a project | asked when a hook is enabled in the Hooks panel | asked when the project opens | project hooks run only when trusted before |
| Theme, accent, sound, motion | yes | accent and sound | by design |
| Data folder | `/settings` | Settings | by design |

## Differences that stay

- Themes and motion are TUI settings; the accent color is a web setting; devices and pairing,
  the Files pane and the Project pane exist only in jin web.
- Headless cannot change settings and does not name sessions.
