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

The server listens on `127.0.0.1` only. Every request needs the cookie of a device. The address jin web prints carries a random start
token (`?token=...`); opening it registers that browser as a device and gives it a cookie that
lasts across restarts. Only a hash of each device token is stored in the database.
Requests with another `Host` header (DNS rebinding) or from another origin are refused.
The agent runs `bash` with your permissions, so do not forward the port.

### Behind a reverse proxy (Tailscale)

`--allow-host NAME` (repeatable) also accepts the host name `NAME` and the `https://NAME`
origin. jin still listens on `127.0.0.1`; a proxy that ends HTTPS must forward to it. The
device cookie is `Secure` on these hosts. Example with Tailscale, after you install it and
sign in on the computer and the phone, and enable HTTPS in the tailnet:

```sh
jin web --no-open --port 7373 --allow-host my-mac.tailnet.ts.net
tailscale serve --bg --https=443 http://127.0.0.1:7373
```

Open `https://my-mac.tailnet.ts.net/?token=...` on the phone, with the start token from the
address jin printed (the QR code below needs `--remote`). Never use `tailscale funnel`: it opens the page to the whole internet.

### `jin web --remote`

Does the steps above for you. It checks that `tailscale` is installed, signed in and has
HTTPS certificates (each case has its own message), allows the computer's tailnet name, runs
`tailscale serve`, prints the https address, and removes the `serve` entry on
exit. On macOS it also runs `caffeinate -i` so the Mac does not sleep while jin web runs; on
other systems keep the computer awake yourself. The machine name goes into the public
certificate log when HTTPS is enabled in a tailnet.

### Pairing a phone

The badge at the top of the page shows `remote off`, `not connected` or `N connected`. Click it
to see a QR code; scan it with the phone camera. The code works once and expires after 5
minutes (the window renews it). The phone becomes a device and stays signed in after a
restart. Settings, Devices lists every device with its state, lets you rename it and revoke it;
a revoked device is locked out at once and its open pages close. The local browser is a device
too.

## On a phone

A window narrower than 700px shows one pane at a time: the focused one. The other open panes
stay and come back on a wide screen; split and close-pane buttons are hidden. The sidebar is
a drawer: the menu button opens it, a tap outside or on a session closes it. Settings, Files,
Project and Tasks fill the screen, and the back gesture closes the drawer or the panel
instead of leaving jin. The `…` menu and dialogs open as sheets from the bottom edge.
Buttons and rows are at least 44px high, inputs are 16px so iOS does not zoom, and the page
follows the visible area above the on-screen keyboard.

## The page

- **Sidebar**: one tree. Each project is a row, named by its path (the home folder is `~`, a long path is cut from the left), with activity dots (blinking while a
  session works, steady for an unread answer, violet for background tasks); its sessions
  are the rows under it, one line each: title and time. The arrow folds a project, the
  project of the focused session starts open. On hover a project row shows three buttons:
  the gear opens its Project pane, the box archives it (it leaves the sidebar, its sessions
  stay; restore it in Settings, Archived projects), and `+` starts a new session in it. `+` next to Projects adds a project directory; the search
  icon searches titles and messages in all projects. The header of each pane shows the
  session title and, under it, the project path of that session.
- **Panes**: up to four panes side by side. A pane is a chat, or one of three panels:
  Settings (the gear), Files (the folder button) and Project (the gear of a project, or the
  path under a chat's title); there is at most one of each. Split from a chat pane's header
  or with `Alt+\`, close with `Alt+W`, focus with `Alt+1`–`Alt+4`. A new panel opens in a
  new pane; with four panes open it replaces the focused one, and a chat that loses its pane
  keeps running and comes back from the sidebar. The last chat cannot be closed. A session
  nobody looks at and that is idle is closed on the server; a working one keeps running.
  The page remembers its layout in the browser: the open panes, their sessions, the focused
  pane, whether the sidebar is open and which projects are folded.
- **Status**: a working session shows a `working` badge in its pane header and an accent
  glow that pulses from the bottom edge, behind the composer. Background tasks add a violet
  `N tasks` badge (the violet is the accent shifted in hue); with no agent work running,
  the glow is violet. When both are active, both badges show and the glow stays in the
  accent color.
- **Chat**: answers stream in as Markdown with highlighted code. Reasoning folds away. A
  tool call shows its summary and, for `bash`, `edit` and `write`, the last eight lines of
  output or the diff; click to see all of it. The chat follows new text at the bottom until
  you scroll up (wheel, trackpad, touch or keys). Then it stays where you read. It follows
  again when you scroll back to the end yourself, when you send a message, or when you open
  another session.
- **Chat details**: Chat details in the composer's menu (the three dots), or `Ctrl+O`,
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

Settings (the gear) is a pane that lists its sections; a click opens one, the arrow in the
header goes back. They are the global settings, shared with the TUI:

- **General**: the accent color (presets or any color; stored as `web.accent`) and the
  notification sound, played by the browser when an answer is done or the agent asks.
- **Tools**, **Models** (the scope of the model picker, with a filter), **Prompts**,
  **Hooks** (the global ones), **System prompt**: edited in the page instead of `$EDITOR`.
  System prompt has a tab for the system, compact and handoff text. Reset to latest from git
  (the second click confirms) downloads the newest default of the open tab from the main
  branch of the jin repository on GitHub and replaces that section of the file, so it needs
  a network and drops unsaved edits.
- **Archived projects**: restore a project that was archived in the sidebar.
- **Data folder**: reset or swap the data directory. jin web stops after the move.

The **Project** pane holds what belongs to one project: its name, whether its hooks may run
(trust), its hooks in `.jin/hooks`, and Archive. When a project with `.jin/hooks` is opened
for the first time, the page asks whether its hooks may run.

## Files

The Files pane shows the project of the focused chat (it changes when you focus another
chat) as a tree. The filter narrows the folders you have opened; the eye shows `.git` and
the files that gitignore rules ignore, which are hidden by default. A click on a file opens
its preview, the arrow goes back:

- code with line numbers and highlighting;
- Markdown as it renders in the chat, with a Source toggle;
- pictures;
- No preview for binary files and files over 1 MB (the agent can still read them).

Files updates itself: once a second the page reads again the folders you have opened and the
file you look at (while the page is visible), and it reads when you open a folder or a file.
A file that disappears says so. Pictures are not read again.

`@ Insert` puts `@path` into the message box of the focused chat. The pane only views files:
it cannot create, rename, delete or edit them, and it cannot leave the project folder.

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
