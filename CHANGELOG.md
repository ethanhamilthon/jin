# Changelog

Notable changes of each release, newest first. Each release is a git tag; the
GitHub release for a tag carries the platform archives and `checksums.txt`.

## v0.8.3 — 2026-10-05

### Changed

- The focused pane has a plain blue border and runs a green glow only while its agent
  works. Idle borders of the other panes use the darker border color of the theme.

## v0.8.2 — 2026-10-05

### Changed

- Each pane shows its own todo list and `ask_user` question at its bottom, also when it is
  not focused. Keys still go to the focused pane.
- Pane borders carry a glow that runs around them: green on the focused pane, blue while
  its session works, purple while it waits on background tasks.
- Esc goes back one panel, for example from a setting to `/settings`; it closes only the
  first panel.

## v0.8.1 — 2026-10-05

### Added

- `/projects` marks each project with a dot, like `/sessions`: green for the open project,
  blinking blue while one of its sessions answers, steady blue for an unread answer and
  blinking purple while only background tasks run.

### Fixed

- The `read` line in the system prompt now says that pictures (png, jpeg, gif, webp, bmp)
  can be opened, so models look at screenshots instead of assuming they cannot.
- A session started in the home directory no longer loads every global hook twice: the
  project hook folder `~/.jin/hooks` is the global one, so it is skipped as a project folder.

### Changed

- Files over the 100-line limit are split; no behavior change.
- Docs: the benchmark index and the install example name the current versions.

## v0.8.0 — 2026-10-05

### Added

- **Projects.** Every session belongs to a registered working directory.
  `/projects` lists them, `Enter` switches, `a` adds an existing folder and `d` removes
  one from the list (removing keeps its sessions in the database). The launch directory is
  registered on start. Existing sessions migrate to a project on the first start of this
  version; the migration is idempotent and does not touch sessions or messages.
- **Split panes.** Up to four panes share one input and status area; each pane keeps its
  own session, draft and scroll position. `/vertical` and `/horizontal` split the focused
  pane, `/quit` closes it. `Tab`, `Alt+arrow` or a click moves the focus, and the focused
  pane's `project · session` title sits on the theme's primary color.
- **`/settings`.** One list for the global settings: Sound, Swap config, Reset, Change
  editor, Tools, Scoped models, Motion, Prompts, Hooks, System prompt.
- **`$` shell input.** A draft that starts with `$` runs as a shell command; `/tui` stays
  for interactive full-screen programs.
- **`/reload`.** Rebuilds the system, compact and handoff prompts of a session in place,
  through a serialized core call so it never races a running request.
- Folder completion in the project field: typing lists directories, `↑` `↓` choose one,
  `Tab` completes, `Enter` confirms the path.
- `Space` switches the highlighted prompt or hook on and off.
- The `read` tool returns pictures as attachments, with format and size checks.
- File tools and bash resolve paths against the session's own directory; nested
  `AGENTS.md` files are sent as the agent works in subdirectories.
- Background tasks are adopted per directory, and `jin async run` carries the working
  directory with the task.

### Changed

- The status bar is one line: model and effort on the left, usage and cost on the right.
  The session title and directory moved to the pane header.
- Panels have no tab bar: sessions, prompts and hooks each open their own list.
- Panel titles no longer repeat key hints; the keys are shown in the hint line above the
  input.
- Key hints in the prompts and hooks panels are ordered `/ search` first, and the folder
  field, the sound panel and the tools panel describe themselves the same way.
- The focus order of the command list changed to match the smaller set of commands.

### Removed

- Voice input: recording, transcription, `/voice`, `/voice-provider` and the fal.run
  integration, together with the `malgo` dependency. Release builds use `CGO_ENABLED=0`,
  so no C toolchain or zig is needed any more.
- `/copy`, `/diff`, `/links`, `/todo` and the `Ctrl+T` todo editor. Todos are written by
  the model; the pinned block is read-only.
- The commands merged into `/settings`: `/prompts`, `/hooks`, `/motion`, `/sound`,
  `/tools`, `/change-editor`, `/system-prompt`, `/reset`, `/swap-config` and `/scope`.

### Renamed

- `/async-tasks` → `/tasks`, `/quit-all` → `/qa`, `/split-vertical` → `/vertical`,
  `/split-horizontal` → `/horizontal`.

### Notes

- The projects migration was checked against a copy of a production database (108
  sessions, 10 742 messages): every row kept, all sessions linked to a project, no
  orphans, and a second open changes nothing.
- Deleting the commands above left a few unreachable helpers in the tree
  (`store.AllChanges`, `internal/ui/diff_text.go`, the todo "user edited" plumbing).
  They are not called; cleanup is a follow-up.
- `make check` and the UI and core race tests pass. The TUI was exercised through the
  scripted demo capture and the golden screens; a human pass through the new panes and
  the projects list is still worth doing.
