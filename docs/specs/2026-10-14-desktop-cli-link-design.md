# Desktop app ships the jin CLI

## Problem

The agent runs `jin ...` through bash (sub-agents, `jin docs`, `jin -p`). The desktop
binary only opens the window and starts the daemon (`--internal-daemon`). A user who installs
only the desktop app has no `jin` command. `internal/startup/env.go` puts the folder of the
running binary first in PATH, but for the desktop that folder (`Jin.app/Contents/MacOS`)
holds no `jin`.

## Decisions

- The desktop package carries a second binary: the `jin` CLI of the same version.
- On every start of a release build, the app links `~/.local/bin/jin` to the bundled CLI.
- An existing file or other link at that path is replaced.
- `jin update` refuses to run when the CLI belongs to the desktop app.
- The agent PATH also gets the folder of the bundled CLI, so the agent works without the link.
- If `~/.local/bin` is not in the PATH of the login shell, a banner in the web UI says so.

## Out of scope

- Desktop app icon (separate task).
- Self-update of the desktop app.
- Native Windows.

## Packaging

- macOS: the CLI is `Jin.app/Contents/Resources/jin`. It cannot live in `Contents/MacOS`:
  `Jin` and `jin` collide on a case-insensitive file system.
- Linux: the archive holds `jin-desktop` and `jin` side by side.
- `scripts/build-desktop.sh` also builds the CLI (`CGO_ENABLED=0`, same `-ldflags` and
  version as `scripts/build-release.sh`) and puts it into the package before `codesign`.
- `scripts/build-release.sh` and the standalone CLI archives do not change.
- `.github/workflows/release.yml`: no new job. The `desktop` jobs already run
  `build-desktop.sh`; they only need the web UI, which `make web` builds.
- `make desktop` (dev) builds no bundled CLI.

## Agent PATH

The desktop binary calls `bundled.AddToPath` at start (`cmd/jin-desktop/main.go`), in the window
and in the daemon. Its folder then reaches the bash tool, tasks and prompt commands, because they
inherit the PATH of the process. `internal/startup/env.go` does not change.

## Link (`internal/desktop/cli_link.go`, build tag `desktop`)

1. Find the bundled CLI: macOS `../Resources/jin` from `os.Executable()`, Linux `jin` next to
   `jin-desktop`.
2. If the file does not exist (dev build), do nothing. This is the only dev guard.
3. Create `~/.local/bin` (`0755`) if needed.
4. Create a symlink under a temporary name, then `rename` it over `jin`.
5. If the link already points to the bundled CLI, change nothing.
6. Run in the background from `startup` in `app.go`, before `connect`. An error does not stop
   the window.

PATH check: run `$SHELL -lc 'echo $PATH'` with a 3 second timeout, and test whether
`~/.local/bin` is in the result. On error or timeout, show no PATH notice. A direct
`os.Getenv("PATH")` is wrong: an app started from Finder gets a short PATH.

## `jin update`

`internal/update/cli.go` resolves the real path of the binary. If that path is inside a
`.app` bundle (macOS) or next to `jin-desktop` (Linux), `jin update` exits with
`This jin belongs to the Jin desktop app. Download the new app from the release page.` and
changes nothing. This applies to every form of `jin update`, including `--check`.

## Notice route

- `GET /desktop/notice` is served by the app itself (`app.go`), not by the daemon proxy. The
  fixed `/api` list in `routes.go` does not change.
- Response `204` when there is nothing to report. Otherwise JSON:
  `{"kind":"path","dir":"~/.local/bin","line":"export PATH=\"$HOME/.local/bin:$PATH\""}` or
  `{"kind":"link","message":"..."}` when the link cannot be written.
- The handler waits up to 4 seconds for the background check, then answers.
- Other methods get `405`. The origin check is the same as the proxy one (`shellOriginAllowed`).
- In plain `jin web` the path answers `404`, so no banner shows.

## Web UI

- `web/src/lib/desktop-notice.svelte.ts` requests the route once at load.
- `web/src/components/DesktopNotice.svelte` shows the text, a Copy button for the `export`
  line, and a close button. `App.svelte` mounts it.
- Closing hides the banner for the current run only. It shows again at the next start until
  the problem is fixed. Nothing is stored.
- The banner is fixed at the bottom left of the page.
- All text is English.

## Tests

- Go, link: create, replace a regular file, replace a foreign link, repeat without change,
  skip without a bundled CLI, error on an unwritable folder.
- Go, route: `204`, JSON, `405`, waits for the background check, refuses a bad origin.
- Go, update: refusal inside a bundle path.
- Go, env: bundled folder in PATH only when the bundled CLI exists.
- Web (vitest): no banner on `404` and `204`; banner on JSON; close hides it for the run.
- Manual: build `make desktop-release` for the local target, open the package, check that
  `jin --version` from a new terminal matches the app, and that the daemon accepts both.

## Risks

- `jin update` and the link replace a CLI installed by `install.sh`. This is intended; the
  user can install it again.
- If the user moves the app, the link breaks until the next start. The next start repairs it.
- The package grows by the size of the CLI binary.
