# Plan: jin web

Status: draft. The look comes from a design.md file that the user will provide; nothing
visual is decided here.

## Decisions taken

- `jin web` is a subcommand of the same binary. It serves on `127.0.0.1` only, with no
  login, and opens the browser.
- Frontend: a component framework (React or Svelte, see open decisions) built with Vite.
  The built files are embedded into the binary with `embed`.
- Transport: Server-Sent Events from server to browser, plain JSON `POST` from browser to
  server. No new Go dependencies.
- v1 aims at full parity with the TUI.

## Architecture

```
main.go              cli.Web -> web.Run
internal/cli         new kind Web for "jin web [--port N] [--no-open]"
internal/session     (new) session runtime shared by ui and web, moved out of internal/ui
internal/web         HTTP server, API handlers, SSE hub, embedded dist
web/                 frontend source (Vite project), built into internal/web/dist
```

Layers stay as AGENTS.md asks: `web/` is UI, `internal/web` is a thin adapter,
`internal/session` + `internal/core` are Core, `internal/provider` is Provider.

### Step zero: a shared session runtime

Today `internal/ui` owns the logic that is not about drawing: starting an agent for a
session (`app_run.go`, `app_sessions.go`, `app_updates.go`), session ownership, provider
switching, model cycling, `$` shell runs (`bash_run.go`, `bash_claim.go`), undo, hooks
trust, compaction triggers, tasks. That logic moves into `internal/session` behind a small
API (open, send, stop, answer, undo, switch model, ...) that emits typed events. The TUI is
rewired onto it first, with the golden tests proving nothing changed. Only then does the web
adapter use the same API. This avoids two diverging copies of the agent wiring.

### Server (`internal/web`)

- `GET /` and assets: embedded `dist`.
- `GET /api/events`: one SSE stream per browser tab; every event carries a session id.
  Events are the JSON form of `core.Update` plus session, task and project events.
  Reconnect resumes with `Last-Event-ID` from a bounded ring buffer; older gaps reload the
  session history.
- `POST /api/...`: commands, one handler file per area (sessions, projects, providers,
  models, prompts, hooks, tasks, settings, files, undo, shell).
- Safety, since the agent has unrestricted `bash`: bind to loopback only, check the `Host`
  and `Origin` headers (blocks DNS rebinding and other sites' pages), require a random
  token generated at start and passed in the opened URL, kept in a cookie.
- The data directory lock (`datadir.Hold`) and session ownership (`store_owner.go`) work as
  in the TUI, so `jin` and `jin web` can run side by side; a session owned by another
  process opens read-only.

### Frontend (`web/`)

- One store per open session fed by the SSE stream; components per TUI feature.
- Files stay around 100 lines, the same rule as Go code.
- Markdown and diff rendering, code highlighting: library choice needs approval.

## Feature parity map

| TUI | Web |
| --- | --- |
| work view, streaming, reasoning, tool folds, diffs | chat view with collapsible tool blocks |
| up to 4 panes, shared input | split layout with up to 4 panes, one focused input |
| slash commands and panels (`/sessions`, `/projects`, `/provider`, `/model`, `/theme`, `/settings`, `/tasks`, hooks, prompts) | command palette (`/`) plus the same panels as side sheets or dialogs |
| `#prompt` and `@file` completion | the same completion popups in the input |
| `ask_user`, todos, `tell_user`, suggestions | the same blocks per pane |
| `$` shell commands | output streamed into the chat |
| `/tui` full-screen programs | open decision (needs a terminal emulator) |
| image paste, clipboard | browser paste and drag and drop |
| undo / rewind | same actions in a message menu |
| themes | CSS variables from design.md |
| sound | Web Audio |
| first-run provider setup | onboarding screen |
| external `$EDITOR` for prompts and hooks | in-browser editor |
| Ctrl+C stop, Ctrl+M model, Ctrl+O folding | the same shortcuts where the browser allows them |

## Phases (one commit or more per finished feature)

1. Extract `internal/session`; TUI uses it; `make check` and golden tests stay green.
2. `jin web` command, server skeleton, token/Host/Origin checks, SSE hub, embedded dist,
   Makefile and release script build the frontend.
3. Chat in one session: send, stream, stop, tool blocks, diffs, history load, usage bar.
4. Sessions and projects lists, ownership and read-only mode, onboarding and providers,
   model and effort.
5. ask_user, todos, tell/suggest, compaction and handoff notices, retries and errors.
6. Panes, command palette, `#prompts`, `@files`, image paste, `$` shell.
7. Tasks, hooks (with trust), prompts editor, settings, data folder actions, undo.
8. Themes and polish from design.md, sound, `/tui` if approved.
9. Docs: `docs/web.md`, README section, `how-it-works.md` layout update, CHANGELOG.

## Open decisions (need the user's answer)

1. React or Svelte (and TypeScript or not).
2. Committed `internal/web/dist` (plain `go build` keeps working, larger diffs) or built in
   CI and by `make` only (needs Node for every source build).
3. `/tui` in the browser: skip, or xterm.js plus a PTY library (a new Go dependency).
4. Libraries for Markdown, highlighting and diffs in the frontend.
5. Default port and whether `jin web` opens the browser by default.
6. Should a browser tab and a running TUI see each other's live sessions (only read-only
   today), or is the current ownership model enough.
