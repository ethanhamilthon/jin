# Changelog

Notable changes of each release, newest first. Each release is a git tag; the
GitHub release for a tag carries the platform archives and `checksums.txt`.

## v0.9.5 — 2026-10-09

### Added

- The pinned todo list folds to its `todo N/M` line: `Ctrl+T` or a click on that line in the
  TUI, a click on the `[ todo ]` header in jin web. The state is kept per session in memory.
- The `#subagents` prompt is now an example in the gallery (`docs/gallery.md`).

### Changed

- jin web status display: the focused pane has an accent border (with two or more panes). A
  working session shows its badge and an accent glow that pulses from the bottom edge. Background
  tasks show a violet badge, and a violet glow when the agent is not working. The violet is the
  accent shifted in hue and needs relative color syntax (Chrome 119+, Safari 16.4+, Firefox 128+).
- jin web suggestion: it shows in the empty input in the accent color. `Space` puts it into the
  input to edit, `Enter` sends it, like the TUI.
- Error text of the nesting limit is `depth limit`; prompts and docs no longer mention sub-agents.

### Removed

- The built-in `#subagents` prompt. Copy it from the gallery to `~/.jin/prompts/subagents.md`
  to keep using it; an existing file of that name now works as an ordinary prompt.

### Fixed

- jin web resyncs slow pages instead of dropping the event stream.

## v0.9.4 — 2026-10-08

### Changed

- README: the documentation index now includes the browser UI (`jin web`).
- About: clarify that the terminal and browser interfaces share the same data and that
  `jin web` runs locally on `127.0.0.1`.
- Database docs: document the `web.accent` setting and its default.
- This is a documentation patch; all features from v0.9.3 are included unchanged.

## v0.9.3 — 2026-10-07

### Added

- Pictures that the `read` tool returns show in the chat: in jin web as an image (click
  opens it in a new tab), in the TUI as a small picture made of half blocks.
- Markdown images open on click: in jin web in a new tab (local paths are served, but only
  files that really are pictures), in the TUI in the system viewer or the browser.
- jin web Markdown: `> [!NOTE]` style alerts, footnotes, `==highlight==`, styled task lists,
  `<details>`, `<kbd>`, `<sub>`, `<sup>`, definition lists and table alignment.
- jin web: the paperclip in the composer attaches files. Pictures go to the model as images,
  other files are saved under the data folder and listed in an `<attached-files>` block.
- jin web: the three dots in the composer open Show tool outputs, Context, Compact, Handoff,
  Rewind, Undo, Reload prompts and Export as Markdown. The Context window also shows
  spent, input and output.
- jin web: the header shows the project name and its path.

### Changed

- jin web: pictures you attach are sent to the model as separate image parts, not as
  `[image 01: /path]` text, and show as thumbnails in your message. A model without vision
  gets no pictures. The `read` tool is unchanged.
- jin web: sessions are children of their project in one sidebar tree; the arrow folds a
  project and `+` on its row starts a session in it. Search covers all projects.
- jin web: the model picker sits next to Send, Stop is an icon, and the composer has no
  border or shadow.
- jin web: the usage numbers moved from the line under the composer into the Context window.

### Removed

- jin web: `/` command completion, the command palette (`Ctrl+K`) and the Commands button.
  Every command is a button or a menu item now.
- jin web: the project badge next to the session title.

## v0.9.2 — 2026-10-07

### Changed

- jin web: the chat shows tool calls or hides them (`tools show / hide`, in the line under
  the composer next to the usage) instead of four folding modes; `Ctrl+O` toggles.
- jin web: the composer floats over the chat with no background band behind it.
- jin web: scroll bars are hidden; everything still scrolls.
- jin web: several `ask_user` questions come one at a time with Next and Back.

### Fixed

- jin web: long lines and wide code blocks no longer stretch messages past the column;
  code scrolls inside its block.

## v0.9.1 — 2026-10-07

### Added

- jin web: the four folding modes of the TUI as tabs in the bottom right corner of the
  composer (Output, All, No tools, Messages); `Ctrl+O` steps through them. The choice is
  saved and shared with the TUI.

### Changed

- jin web: a pasted or dropped picture shows as a thumbnail above the composer instead of
  an `[image 01]` token in the text; a message may be pictures only.
- jin web: the composer is two lines high by default.
- jin web: the header no longer shows how many sessions work.
- jin web sidebar: wider; sessions show only the selected project (the "all projects"
  toggle is gone); a search icon and a `+` next to Sessions replace the search field and
  the New button; each session is one line with its title and time.

### Fixed

- A message sent while the agent works no longer leaves the session busy after the
  answer, in jin web (the working pill and Stop stayed) and in the TUI (the pane kept
  glowing and the session stayed claimed). The agent now reports such a message with
  `UpdateTaken`.

## v0.9.0 — 2026-10-07

### Added

- `jin web [--port N] [--no-open] [--cwd DIR]`: jin in the browser, on `127.0.0.1:7373`
  (or the next free port), guarded by a token. Same data as the TUI: projects, sessions,
  providers, prompts, hooks, settings and tasks. Up to four panes, a command palette,
  `/`, `#`, `@` and `$` in the composer, pictures by paste or drop, and settings edited in
  the page with a user-set accent color. See docs/web.md.

### Changed

- Rebuilding a session's entries from its history, tool output lines, rewind points,
  scope filtering and provider names moved to `internal/session`, which the TUI and
  jin web share.
- Source builds need Node 22 for the web UI; without it the UI is left out.

## v0.8.4 — 2026-10-05

### Added

- `task` tool: start a command in the background (`dir`, optional `stdin`), then `check`,
  `input`, `stop` or `list`. Results arrive as messages by themselves.
- When the agent answers while its tasks still run, it gets one note listing them, stops
  what it no longer needs and adds one line to its answer.
- `tell_user` tool: `message` shows a line in the chat without ending the turn; `suggest`
  offers your next request, greyed in the empty input (`Enter` sends it, `→` edits it).
  `jin -p` adds it to the result as `suggestion`.
- `bash` takes `dir` to run a command in a subdirectory.

### Changed

- Every tool call of one answer runs at the same time, `bash` included. Only `ask_user`,
  todo updates and a second call on a file another call of the answer changes wait.
- Background tasks belong to the jin process that started them and stop when it exits.
  `jin async`, `jin daemon` and the async tables are gone; logs moved to `~/.jin/tasks`.
  `#subagents` starts sub-agents as tasks; they report their final answer only.
- A non-zero `bash` exit reads `[exit code: N]` instead of `[command failed: ...]`.

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
