# Plan: jin web

Status: draft. The look is defined in [web-design.md](web-design.md).

## Decisions taken

- `jin web` is a subcommand of the same binary. It serves on `127.0.0.1` only, with no
  login, and opens the browser at once.
- Frontend: Svelte 5 with TypeScript, built with Vite. Chosen over React for speed and a
  smaller bundle. The built files are embedded into the binary with `embed`.
- Only frontend source is committed. `internal/web/dist` is git-ignored and built by
  GitHub Actions (`ci.yml`, `release.yml`) and by `make build`; see Build below.
- Frontend libraries: `marked` for Markdown, `DOMPurify` to sanitize its HTML (model
  output is untrusted), `highlight.js` for code. Diffs need no library: the server sends
  the lines that `internal/diff` already computes. Fonts (Source Serif 4, Inter Tight,
  JetBrains Mono) come from `@fontsource` packages and are bundled, so the page works offline.
- No `/tui` in the browser.
- Transport: Server-Sent Events from server to browser, plain JSON `POST` from browser to
  server. No new Go dependencies.
- v1 has everything the TUI can do, but not its form. Panels, key chords, fold modes and
  slash commands exist in the TUI because a terminal has few ways to show things; the web
  UI is designed for the browser (sidebar, tabs, dialogs, hover, drag and drop) by
  web-design.md.
- Same data as the TUI: one data directory, the same SQLite store, projects, sessions,
  providers, prompts, hooks, settings and tasks. A session started in one is visible in
  the other.

## Architecture

```
main.go              cli.Web -> web.Run
internal/cli         new kind Web for "jin web [flags]"
internal/session     (new) session runtime shared by ui and web, moved out of internal/ui
internal/web         HTTP server, API handlers, SSE hub, embedded dist
web/                 frontend source (Vite project), built into internal/web/dist
```

Layers stay as AGENTS.md asks: `web/` is UI, `internal/web` is a thin adapter,
`internal/session` + `internal/core` are Core, `internal/provider` is Provider.

### Command line

```
jin web [--port N] [--no-open] [--cwd DIR]
```

| Flag | Meaning |
| --- | --- |
| `--port N` | listen on this port; busy is an error |
| `--no-open` | print the URL instead of opening the browser |
| `--cwd DIR` | the project the page opens first, default the current directory |

Without `--port`, jin web takes 7373 (unassigned in the IANA registry and not a common
dev server port). When it is busy it tries 7374 to 7383, then any free port, and prints the
URL it got. Flags work before and after each other, like `jin -p`; `jin web --help` lists
them.

### Build

- `web/` is a Vite project; `npm run build` writes `internal/web/dist`.
- `internal/web` embeds `all:dist`; a committed `dist/.gitkeep` keeps `go build` and
  `go test` working without Node. A binary built that way answers `jin web` with
  "this build has no web UI: run make build or install a release".
- `make build` and `make build-prod` run the frontend build first when Node is present;
  `make web-dev` runs Vite's dev server with its API proxy pointed at a running `jin web`.
- `ci.yml` and `release.yml` add `actions/setup-node`, `npm ci`, `npm run check` (svelte-check
  and tests) and `npm run build` before the Go steps, so releases always carry the UI.

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

- One store per open session fed by the SSE stream; components follow web-design.md.
- Files stay around 100 lines, the same rule as Go code.

## Capabilities (what, not how)

Every capability below works on the shared data. How each one looks and is reached comes
from web-design.md, not from the TUI.

- Projects: list with activity, open, switch, per-project session list.
- Sessions: new, open, search, rename, delete, export, read-only when another process
  owns one; several sessions visible at once.
- Chat: send, stream answer and reasoning, stop, queue messages while the agent works,
  tool calls with their output and diffs, usage, context and cost.
- Agent interaction: `ask_user` questions, todos, `tell_user` messages, suggested next
  request, compaction and handoff notices, retry and error lines.
- Input: `#prompt` and `@file` mentions, image paste and drop, `$` shell commands.
- Providers and models: first-run setup, add, edit, delete, default provider, model and
  effort per session.
- Prompts and hooks: list, create, edit in the browser, project hooks trust.
- Background tasks: list, live output, stop.
- Undo and rewind of file changes.
- Settings, accent color, sound, data folder reset and swap, update notice.

TUI-only mechanics left out on purpose: fold modes, `Esc` panel stack, `Ctrl+M` model
cycling, the external `$EDITOR`, `/tui`.

## Phases (one commit or more per finished feature)

1. Extract `internal/session`; TUI uses it; `make check` and golden tests stay green.
2. `jin web` command, server skeleton, token/Host/Origin checks, SSE hub, embedded dist,
   Makefile and release script build the frontend.
3. Chat in one session: send, stream, stop, tool blocks, diffs, history load, usage bar.
4. Sessions and projects lists, ownership and read-only mode, onboarding and providers,
   model and effort.
5. ask_user, todos, tell/suggest, compaction and handoff notices, retries and errors.
6. Several sessions at once, `#prompts`, `@files`, image paste, `$` shell.
7. Tasks, hooks (with trust), prompts editor, settings, data folder actions, undo.
8. Accent setting and polish from web-design.md, sound.
9. Docs: `docs/web.md`, README section, `how-it-works.md` layout update, CHANGELOG.

## Open decisions (need the user's answer)

1. Should a browser tab and a running TUI see each other's live sessions (only read-only
   today), or is the current ownership model enough.
