# Roadmap to jin v1.0.0

Date: 2026-10-09. Current version: v0.9.5.

## Goal

v1.0.0 is a finished jin: stable, feature-complete and published. After 1.0 the data
formats and command line stay compatible. A new feature or breaking change is brainstormed
and approved on its own before it is built; nothing here pre-approves it.

Out of scope for 1.0: MCP client, plugins, permission prompts or a sandbox. The stance in
`docs/extending.md` stays: hooks and plain CLIs.

## How we work

Each workstream is brainstormed when it is its turn. Bounded items get a short design in
chat and are approved before code. Architectural items get their own spec in `docs/specs/`
and a plan. Tick the boxes below as items ship. Every release updates `CHANGELOG.md`.

## Order

| Release | Content |
| --- | --- |
| v0.9.6 | Two web bugs (1) |
| Spike | Windows feasibility (2 days, answer only, throwaway code) |
| v0.10 | Parity across TUI, web and headless (9) |
| v0.11 | WebUI work (3) |
| v0.12 | Remote access and mobile UI (10) |
| v0.13 | Context for big projects (4) |
| v0.14 | Windows support (5) |
| v0.15 | Benchmarks and README refresh (6) |
| Not scheduled | Context transparency (11); the owner picks the release, and it must land before v0.99 because it changes what the model receives |
| v0.99 | Release candidate: contract freeze and docs (7) |
| v1.0.0 | Publish (8) |

Shipped: v0.9.6 (web bugs), the Windows spike (result in section 2) and v0.10.0 (parity, plus
the removal of the `todo` and `tell_user` tools, the macOS and `TODO.md` prompt notices,
system prompt reset buttons and two jin web fixes).
v0.11.0 (WebUI): Markdown look, settings, project and files as panes, archive projects, the
Windows build fix and a kept layout. Left from section 3: nothing. Next: v0.12, remote access
and the mobile UI (section 10).

The Windows spike runs before v0.10 so its result can move the Windows release earlier.
Parity goes first because every later feature that adds a tool or a setting would
otherwise have to be wired into three places. Section numbers are topics, not order.

## 1. Web bugs (v0.9.6)

Bounded. Causes below are from reading the code, not yet reproduced in a browser.

- [x] Top bar shows only `jin`. Decided: remove the project name and path from
  `TopBar.svelte` (it showed `app.project || app.dir`, one global value that ignored the
  focused pane). The project stays visible in the sidebar tree.
- [x] Pane header shows the project. Decided: in `PaneHeader.svelte` the session title gets
  a little smaller (now 20px serif), and under it a smaller line shows the project name and
  path of that pane's own session. The path is cut with an ellipsis in narrow panes and its
  full text is the tooltip.
- [x] Focus follows the input (decided). `app.focused` stops being
  separate state: the pane whose composer has the cursor is active, the last one stays
  active while focus is in the sidebar or a dialog, and a click on a read-only pane still
  activates it. Check that selecting chat text to copy is not broken. Web only: the TUI
  has one shared input, so its focused pane stays as it is.
- [x] The `…` menu in the composer is not clipped in a split. `MoreMenu.svelte` anchors the
  menu to `right: 0` with 220px minimum width, so in a narrow pane it opens past the left
  edge. It must stay inside the viewport.
- [x] The `…` menu is redesigned (moved here from the WebUI section, decided). Every item
  has an icon on the left that matches its meaning. The right side shows its kind: an arrow
  when it opens a window, a run icon when it does something at once (for example Compact),
  a rectangular switch when it is a toggle (Show tool outputs).

## 2. Windows spike

Question: how much work is native Windows, and which shell does the `bash` tool use?
Look at: `bash_other.go` and `lock_other.go` stubs, process kill, `KillBackground`, tcell
on Windows Terminal, `jin web` browser launch. Output: a recommendation and a size
estimate. Anything built is throwaway.

Result (done 2026-10-09, cross-compile and code reading only, nothing run on Windows):

- `GOOS=windows go build ./...` fails in one place: `syscall.Kill` in
  `internal/store/running.go` (`processAlive`). With that line stubbed the whole project
  builds. Tests also use `syscall.Kill` (`internal/ui/bash_run_test.go`).
- Stubs that behave wrongly on Windows: process-tree kill (`tools/bash_other.go`,
  `tasks/other.go`, `dyn/group_other.go`), `KillBackground`, the data-folder lock
  (`datadir/lock_other.go`), and replacing the running binary (`update/install.go`).
- `bash -c` is hard-coded in `tools/bash_run.go`, `tasks/start.go`, `dyn/run.go`; `sh -c` in
  `ui/bash_run.go`, `session/shell.go`, `editor/editor.go`. New code must not add more.
- Browser launch already has a `rundll32` branch. Data lives under the user home folder.
- Not verified: tcell in Windows Terminal.
- Decided: the shell on Windows is Git Bash only; no PowerShell tool. Lookup order as in pi:
  a setting for the path, the standard Git folders, then `bash.exe` on `PATH`; a clear
  error when none is found. Claude Code and pi use Git Bash the same way; PowerShell is a
  separate tool there and can be revisited after 1.0.
- Windows stays at v0.14. Size: about 15 to 20 files.

## 3. WebUI (v0.11)

Each item is bounded unless the brainstorm says otherwise. The look stays as in
`docs/web-design.md`; only the settings window is redesigned, to match the chat window.

- [x] Markdown rendering (done 2026-10-09, mockup `docs/specs/mockups/markdown.html`): typography (line height, line width, spacing), blocks (code,
  tables, quotes, lists) and hierarchy (headings, emphasis, rules, links) in
  `web/src/styles/markdown*.css`. Update `web-design.md` where it changes.
- [x] Settings, Files and Project as panes (decided 2026-10-09). The settings window is no
  dialog any more: it is a pane of the same kind as a chat pane, one of the up to four open
  panes, with the same frame, header, typography, spacing and controls. Today a pane is
  `{key, session}` (`app.svelte.ts`); it becomes a pane with a kind: chat, settings, files or
  project. Decided:
  - Settings: at most one pane. Files: at most one pane. Project: a separate pane kind that
    holds the project settings (candidates: name, hooks trust, default model, archive).
  - Opening: the gear (or the Files or Project button) opens a new pane while fewer than four
    are open. With four open it replaces the focused pane; the chat of that pane keeps living
    and comes back with a click in the sidebar.
  - Settings layout: a list of sections; a click opens the section with a back button, as in
    mobile settings. It works in a narrow pane.
  - Files follows the project of the focused chat pane and changes with it.
  - The pane layout is kept after a page reload (decided): which panes are open, their kinds
    and order, which session each chat pane shows, and the focused pane. Where it is stored
    (browser storage or the settings table, so the TUI and other browsers agree) is settled
    in the brainstorm.
  - Mockups approved 2026-10-09 (static HTML in `docs/specs/mockups/`: `settings-panes.html`,
    `settings-sections.html`, `files-pane.html`). Decisions from them:
    - A section opens with a back arrow only (no text) in its header.
    - Rows are hairline-separated; switches are rectangular and always the last control in a
      row; action buttons (Edit) and badges (`system`, `project`) come before it, and a badge
      follows the name.
    - Hooks in Settings are global only (no "project" checkbox). Project hooks and their trust
      live in the Project pane. (Proposed in the mockup review; the owner did not object.)
    - System prompt: tabs System, Compact, Handoff, one "Reset to latest from git" button for
      the open tab.
    - Models gets a filter field.
    - Files: tree with a filter field and an eye button for hidden files; click opens a
      preview (code with line numbers, Markdown with a Source toggle, picture, or "No
      preview" for binary files and files over 1 MB); `@ Insert` puts the path into the
      focused chat's composer. `.git` and ignored files are hidden by default.
  - Applies to `SettingsDialog.svelte` and all its tabs, `Workspace.svelte`, `Pane.svelte`,
    `PaneHeader.svelte`, `keys.ts`. A mockup comes before code.
- [x] Archive projects (decided): the sidebar can archive a project. The project and its
  sessions stay in the database and on disk; the project is only hidden from the sidebar.
  An archived project is restored in Settings, in a list of archived projects. Needs a flag
  on the project in the store, and the TUI `/projects` list must agree on whether it shows
  archived projects (settled in the brainstorm).
- [x] Windows build fix: replace `syscall.Kill` in `internal/store/running.go`
  (`processAlive`) and in `internal/ui/bash_run_test.go` with build-tagged helpers, so
  `GOOS=windows go build ./...` and `go vet` pass. Small; the rest of Windows stays in v0.14.
- [x] File explorer, view only (a Files pane): project file tree and file preview (code, markdown,
  pictures), insert `@path` into the composer. No create, rename, delete or edit. Needs a
  new read-only API next to `internal/web/api_files.go`.

## 4. Context for big projects (v0.13)

Architectural. Nested `AGENTS.md` already ships. Missing: a search or repo-map
capability; the model uses `bash` today. Brainstorm decides between `grep`/`glob` tools,
a repo map, or both, after measuring where agents lose time on large repos. Tools stay
small and in `internal/tools`.

- [ ] Measure the problem on a large repo.
- [ ] Spec, plan, implementation, docs.

## 5. Windows (v0.14)

Architectural. Known gaps: no windows target in `scripts/build-release.sh`, `install.sh`
accepts only macOS and Linux, no PowerShell installer, `bash` tool assumes POSIX shell,
process-group kill and data-folder lock are stubs. Scope comes from the spike.

- [ ] Spec and plan from the spike result.
- [ ] Shell and process control, file lock, installer, release archive, CI on a Windows
  runner, TUI and web checks, docs.

## 6. Benchmarks (v0.15)

Today: Aider Polyglot, Python only, 20 and 10 tasks, jin v0.7.2. Run again on the release
that has all features.

- [ ] Decide task sets: more languages, larger tasks, more runs per task.
- [ ] Run jin against the same agents as before with `tbench`.
- [ ] Update `docs/benchmarks.md` and the README tables with the version and method.

## 7. Release candidate (v0.99)

- [ ] Decide what is frozen: database schema, hook and prompt file formats, settings keys,
  CLI flags and commands, web API. Document it.
- [ ] Upgrade tests from older data folders (`internal/upgrade`).
- [ ] Sweep known bugs; `make check` and `npm run check` are green.
- [ ] Docs pass: every doc matches the code; the README states what 1.0 promises.
- [ ] Fresh install check on every supported platform.

## 8. v1.0.0

- [ ] Version policy written down (what a major, minor and patch change means).
- [ ] Tag, release archives and checksums, `CHANGELOG.md` entry.
- [ ] Announcement text and README badges or links as needed.

## 9. Parity across TUI, web and headless (v0.10)

Goal: jin is equally strong in all three modes and the agent behaves the same in each.
Interface differences stay (themes, panes, key chords); agent behavior must not differ.

Found in the code: the agent is built in three places (`internal/headless/agent.go`,
`internal/session/start.go`, `internal/ui/app_sessions.go`), so tools, prompts, hooks,
tasks and compaction can drift apart.

- [x] One shared agent factory used by TUI, web and headless.
- [x] A test that builds the agent in all three modes and compares tools, system prompt
  parts, hooks, background-task wiring and compaction settings.
- [x] A parity table (feature by mode) in `docs/`, with every gap either closed or marked
  as intentional.

Headless long sessions (decided: stay one-shot, no long-lived process):

- [x] Command analogues for session control: compact, handoff, rewind, undo, context.
  Names and shape are decided in this workstream's brainstorm.
- [x] A reload command for headless, like `/reload` in the TUI and web.
- [x] Keep the prompt prefix stable across `jin -p -c` runs, so the provider cache stays
  valid. Today `startup.Render` runs on every start, so `{{commands}}` output can change
  the system prompt. Not yet measured; measure first.

Known and accepted: background tasks end with the headless process; there is no
`ask_user` and no streaming of deltas in headless. Revisit only if asked.

Decided in the brainstorm (2026-10-09):

- The factory covers agent construction only: tools, workdir, background wiring, refresher,
  side prompts, context size. The start-up render (`startup.Input`, `finishRender`) stays
  duplicated in `internal/ui` and `internal/session` for now; the TUI is not moved onto
  `session.Manager` in v0.10.
- Headless stays text only: no pictures.
- Headless session control is a `jin session` command group:
  `jin session compact|handoff|rewind|undo|context|reload <id>`. Flags and output are
  settled in the spec of this workstream.

Found while reading the code: `internal/ui` (TUI) and `internal/session` (web) each have
their own start, render and reload code. Headless sets no refresher and no workdir on the
agent and ends background tasks with the run. The spec must mark each such difference as
intentional or close it.

To decide in the spec: which settings and commands from the TUI and web are missing in the
other two.

## 10. Remote access and mobile UI (v0.12)

Goal: work with a jin that runs on your computer from a phone, with little setup.
Architectural, and security-critical: jin has an unrestricted `bash`, so anyone who
reaches the web UI runs commands on the machine.

Today (`internal/web/guard.go`, `listen.go`): listens on `127.0.0.1` over plain `http`;
only the hosts `127.0.0.1:port` and `localhost:port` are accepted; a random token that
changes on every start comes in the URL once and then in a cookie. The layout is built
for a desktop.

Decided:

- Tailscale is the only remote path for now. Other providers (Cloudflare Quick Tunnel,
  others) come later; at most a documentation recipe.
- Tailscale is not a dependency. The user installs it and signs in on the computer and the
  phone. Jin only detects the `tailscale` command and runs it; it embeds nothing.
- Jin stays on `127.0.0.1`. `tailscale serve` terminates HTTPS and proxies to it. Jin
  never uses `tailscale funnel` (public internet).

Items:

- [x] `--allow-host`: accept an additional host name and the `https` origin in `guard.go`.
  Document the manual `tailscale serve` recipe.
- [x] `jin web --remote` (QR by `skip2/go-qrcode`, `caffeinate` on macOS, decided
  2026-10-09; only the signed-out message is checked against a real Tailscale): check that `tailscale` is installed and signed in, run
  `tailscale serve`, wait until it is ready, print the address and a QR code. Clear
  messages when Tailscale is missing, signed out, or HTTPS is not enabled in the tailnet.
- [x] (done 2026-10-09: badge in the top bar opens a QR code with a one-time code of 5
  minutes; the phone becomes a device kept in the `devices` table; Settings, Devices lists,
  renames and revokes; the terminal QR was removed) Pairing that survives a restart, so the phone does not need a new QR every day.
  Stronger than today's per-start token; shape decided in the brainstorm.
- [x] (done 2026-10-09, mockup `docs/specs/mockups/mobile.html`; checked in a 390px browser
  window, not yet on a real phone) Mobile layout of the web UI: one pane, sidebar as a drawer, touch targets, composer
  with the on-screen keyboard, dialogs that fit a small screen.
- [ ] Docs: setup guide, security notes (the machine name goes to the public certificate
  log when HTTPS is enabled; the computer must stay awake and `jin web` must run).

Open:

- The QR code needs a library or our own code. A new dependency needs the owner's
  confirmation (AGENTS.md rule 4); decide before building.
- Shape of pairing (one-time code, passcode, device list with revoke).
- Whether `--remote` also keeps the computer awake.

## 11. Context transparency (not scheduled)

Status 2026-10-09: the owner postponed this section, `jin docs` included. Work on it starts
later; nothing here is scheduled.

Goal (decided): the owner and the user can see 100% of what the model receives. The context
is made of the system prompt and the hooks, and nothing else. Code does not add text of its
own to what the model reads.

Found in the code on 2026-10-09 (what jin adds today besides the system prompt and hooks):

In the system prompt (`internal/core/prompt.go`, `prompt_notes.go`, `prompt_env.go`):

- The jin docs pointer (`docs_prompt.md`), always added.
- The `Tools: none` notice when every tool is off.
- Two notices from v0.10: the macOS `sed -i` rule, and the `TODO.md` checklist hint when
  `write` or `edit` is on. They are code-added text, so they conflict with this goal.
- The `AGENTS.md` files of the working directory and its parents.
- The environment block (working directory, OS, date) and the session id.

In messages (`internal/core`, `internal/files`, `internal/prompts`, `internal/tasks`):

- Notes in front of a user message: `<system-refreshed>`, `<files-undone>`.
- Wrappers around user input: `<pasted-prompts>` for `#prompts`, `<attached-files>` for
  attached files, the picture labels.
- Task messages: `<task-result ...>`, and `<background-tasks>` before the agent finishes a
  turn.
- Text added to tool results: the AGENTS.md of a folder appended to `read`, `edit` and
  `write` results (`agentsmd_nested.go`), `[exit code: N]`, truncation notes, the omitted
  note when old tool results are pruned (`prune.go`).
- The compaction and handoff prompts (editable in `system-prompt.md`) and the summary
  message that replaces the history.
- Tool schemas and descriptions (`internal/tools`). They must exist for tool calls; whether
  they count as part of the context here is an open question.

Direction (owner's proposal, 2026-10-09; to be confirmed in the brainstorm): everything that
code adds to the system prompt moves into the default `system-prompt.md` as `{{commands}}`,
so the user can read and change any of it. The sketch:

- `jin docs` (decided): all of jin's documentation is built into the binary and read through
  this command, and the agent finds things only through it. It works without internet and
  always matches the installed version, never the main branch on GitHub. `{{jin docs}}`
  prints the short pointer text; `jin docs <topic>` prints one doc page. Search and topic
  list shape are settled in its brainstorm.
- `{{jin tooldefs}}` prints a usage guide for every enabled tool.
- Environment, session id and `AGENTS.md` become commands too (for example `date`,
  `uname`, `$JIN_SESSION_ID`, a `jin agents` command).

Points to settle for that direction:

- `jin` must be found by the prompt commands. They run through `bash -c` with the
  environment of the process; a dev build or a second install can be a different version, so
  the running binary's folder should come first in `PATH`.
- Session id and directory reach the commands through `startup.Input.Env` (for example
  `JIN_SESSION_ID`, `JIN_DIR`); that field already exists.
- The cache break needs a visible marker line in the file, because the stable part and the
  live part must stay apart.
- Tool schemas are still sent in the `tools` field of the request; the API needs them. A
  `tooldefs` guide in the prompt would repeat them. Choose one source for both so they cannot
  drift, or keep the guide short.
- A user with a custom `system-prompt.md` has none of the new placeholders after an upgrade.
  Options: a migration that adds them, or the web reset buttons plus a note in the release
  notes.
- A command that fails or runs past 10 seconds leaves an error text in the prompt, so the
  model would lose the tool guide or the docs pointer. Decide the fallback.
- Message-level additions (notes, wrappers, text appended to tool results) are not covered by
  this move and need their own decision.

Items:

- [ ] Decide for each line above: remove, move into a hook or the system prompt (so the user
  sees and edits it), or keep and show. The decision is the owner's, one line at a time.
- [ ] A single exact view of what is sent: `/context` and the web Context window show the
  full text of every part, not only sizes, and the same text that goes on the wire.
- [ ] A test that fails when code adds text that is not in the system prompt, a hook or a
  tool schema.
- [ ] Docs: say exactly what the model receives, in `docs/how-it-works.md`.

Open:

- Is `AGENTS.md` loading kept? It is a feature users rely on; the options are to keep it and
  show it as its own labeled part, or to turn it into a hook the user installs.
- The environment block and the session id: kept as shown text, or moved into a default
  hook.
- Pruning and compaction change the history the model sees; they stay, but each must be
  visible in `/context`.
- Which release it ships in.
