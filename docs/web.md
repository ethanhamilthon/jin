# jin web

`jin web` opens jin in the browser. It runs the same agent as the TUI and uses the same
data directory: projects, sessions, providers, prompts, hooks, settings and background
tasks are shared. A session you start in one shows up in the other.

```sh
jin web                 # serve on 127.0.0.1:7373 and open the browser
jin web --port 8080     # this port; a busy port is an error
jin web --no-open       # print the address instead of opening a browser
jin web --cwd ~/code/x  # the project the page opens first
```

Without `--port`, jin web takes 7373. When it is busy it tries 7374 to 7383, then any
free port. It prints the address it got. `Ctrl+C` stops it; running requests stop too and
their sessions are marked unread.

## Access

The server listens on `127.0.0.1` only. Every request needs a random token that jin web
makes at start. The printed address carries it once (`?token=...`), and a cookie keeps it.
Requests with another `Host` header (DNS rebinding) or from another origin are refused.
The agent runs `bash` with your permissions, so do not forward the port.

## The page

- **Sidebar**: one tree. Each project is a row, named by its path (the home folder is `~`, a long path is cut from the left), with activity dots (blinking while a
  session works, steady for an unread answer, violet for background tasks); its sessions
  are the rows under it, one line each: title and time. The arrow folds a project, the
  project of the focused session starts open, and `+` on a project row (shown on hover)
  starts a new session in it. `+` next to Projects adds a project directory; the search
  icon searches titles and messages in all projects. The header of each pane shows the
  session title and, under it, the project path of that session.
- **Panes**: up to four sessions side by side. Split from a pane's header or with `Alt+\`,
  close with `Alt+W`, focus with `Alt+1`–`Alt+4`. A session nobody looks at and that is
  idle is closed on the server; a working one keeps running. With two or more panes, the
  focused one has a border in the accent color.
- **Status**: a working session shows a `working` badge in its pane header and an accent
  glow that pulses from the bottom edge, behind the composer. Background tasks add a violet
  `N tasks` badge (the violet is the accent shifted in hue); with no agent work running,
  the glow is violet. When both are active, both badges show and the glow stays in the
  accent color.
- **Chat**: answers stream in as Markdown with highlighted code. Reasoning folds away. A
  tool call shows its summary and, for `bash`, `edit` and `write`, the last eight lines of
  output or the diff; click to see all of it. The chat follows new text at the bottom until
  you scroll up (wheel, trackpad, touch or keys). After that it never moves by itself, also
  when you scroll back to the end, until you send a message or open another session.
- **Tool outputs**: Show tool outputs and reasoning in the composer's menu (the three dots), or `Ctrl+O`,
  shows or hides tool calls, their output and the model's reasoning. Hidden, the chat keeps only your
  messages, the agent's text between tool calls and its final answer. The choice is saved and shared with the TUI
  (show is its Output mode, hide its No tools mode).
- **Composer**: two lines high, growing with the text. A click anywhere in it, outside its
  buttons, puts the cursor in the input, at the end of the text. `Enter` sends, `Shift+Enter`
  adds a line. Messages typed while the agent
  works are queued. `#` completes prompts, `@` completes paths and attaches files, `$` runs
  a shell command in the project directory (its output stays out of the conversation).
  Paste, drop or pick (the paperclip) a picture to attach it: it shows as a thumbnail above
  the composer (`✕` removes it) and goes with the next message as an image. Other files
  are saved under the data folder and listed in an `<attached-files>` block.
  `Ctrl+C` with no selection, or Stop, interrupts the request.
- **ask_user** questions appear above the composer one at a time, with their options and a
  free answer; Next goes on, Back returns, the last one sends all answers.
- **Composer buttons**: the paperclip attaches files; the three dots open Show tool
  outputs, Context (what fills the window, spent, input and output, cache), Compact,
  Handoff, Rewind, Undo, Reload prompts and Export as Markdown (a download, like
  `jin export --md`). Each item has an icon; the right side shows what it does: an arrow
  opens a window, a triangle runs at once, a rectangular switch is a toggle. The menu
  stays inside its pane. The model button and Stop sit next to Send. The page has no slash
  commands and no command palette: new session, split, close, project, settings, providers
  and tasks are buttons.

## Providers and models

Until a provider is set up the page shows the first-run screen: choose the API kind, give
the name, base URL and key, then the model and effort. Providers (the model picker)
adds, deletes and switches the default provider. The model button in the composer picks
the model and effort of that session; for a session of the default provider it also
becomes the default, as in the TUI.

## Settings

Settings (the gear) holds the global settings, shared with the TUI:

- **General**: the accent color (presets or any color; stored as `web.accent`) and the
  notification sound, played by the browser when an answer is done or the agent asks.
- **Tools**, **Models** (the scope of the model picker), **Prompts**, **Hooks** (with
  project hook trust), **System prompt**: edited in the page instead of `$EDITOR`. The reset buttons (system, compact, handoff or all three; the second click confirms) download the newest default from the main branch of the jin repository on GitHub and replace that section of the file, so they need a network and drop unsaved edits.
- **Data folder**: reset or swap the data directory. jin web stops after the move.

When a project with `.jin/hooks` is opened for the first time, the page asks whether its
hooks may run.

## Differences from the TUI

The `Esc` panel stack, `Ctrl+M` model cycling, the external editor and `/tui` are
terminal mechanics and have no web counterpart. Only one jin process works in
a session at a time: a session owned by a running TUI opens read-only in the page, and
the other way round.

## Building

Release archives carry the UI. A source build needs Node 22 for it: `make build` runs
`npm ci` and `npm run build` in `web/`, which writes `internal/web/dist` (git-ignored).
Without npm the binary builds anyway and `jin web` says that the UI is missing.

For UI work, start `jin web --no-open`, run `make web-dev` and open the Vite address
with the token: `http://localhost:5173/?token=<token>`. `JIN_WEB` points the dev proxy
at another address. `make web-check` runs svelte-check and the unit tests.
