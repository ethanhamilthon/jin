# Plan: desktop app ships the jin CLI

Spec: `2026-10-14-desktop-cli-link-design.md`. Each step is one commit, tests first (TDD).
Files stay near 100 lines. English only.

## 1. Bundled CLI path helper

- New `internal/bundled/bundled.go` (no build tag): `Path(exe string) (string, bool)`.
  macOS: `<exe dir>/../Resources/jin`. Linux: `jin` next to `jin-desktop`. `ok` is true only
  if the file exists. `Inside(real string) bool` tells if a path is a bundled CLI.
- Test `bundled_test.go`: macOS layout, Linux layout, missing file, `Inside` cases.

## 2. Agent PATH

- `internal/startup/env.go`: when `bundled.Path(os.Executable())` exists, add its folder to PATH
  after the folder of the running binary.
- Test in `startup_test.go`: folder present only when the bundled CLI exists.

## 3. `jin update` refusal

- `internal/update/cli.go`: after `executable()`, if `bundled.Inside(target)`, print
  `This jin belongs to the Jin desktop app. Download the new app from the release page.`
  to stderr and return 1. The check goes before `Latest`, so `--check` also refuses.
- Test in `update_test.go`: refusal and no file change.

## 4. Link

- New `internal/desktop/cli_link.go` (tag `desktop`): `installLink(exe, home string) error`.
  Temporary symlink, then `rename`. No change if the link is already correct. Skip when
  no bundled CLI.
- New `internal/desktop/cli_path.go` (tag `desktop`): `inLoginPath(dir string) (ok, known bool)`
  runs `$SHELL -lc 'echo $PATH'` with a 3 second timeout.
- Tests: create, replace file, replace foreign link, repeat, skip, unwritable folder; PATH
  check with a fake shell script.

## 5. Notice route

- New `internal/desktop/notice.go` (tag `desktop`): `notice` state with a ready channel,
  `ServeHTTP` for `GET /desktop/notice` (`204`, JSON, `405`, 4 second wait, origin check).
- `internal/desktop/app.go`: handler branch for `/desktop/notice`. `startup` runs
  `installLink` and the PATH check in the background, then sets the notice.
- Tests: `204`, both JSON kinds, `405`, wait for the check, bad origin.

## 6. Banner

- `web/src/lib/desktop-notice.svelte.ts` and `web/src/components/DesktopNotice.svelte`,
  mounted in `web/src/App.svelte`. Copy button, close button, no storage.
- Vitest: `404` and `204` give no banner, JSON gives a banner, close hides it.
- Run `make web-check`.

## 7. Packaging

- `scripts/build-desktop.sh`: build the CLI (`CGO_ENABLED=0`, same `-ldflags` and version
  as `build-release.sh`). macOS: into `Jin.app/Contents/Resources/jin` before `codesign`.
  Linux: next to `jin-desktop`, both in the tar.
- `.github/workflows/release.yml`: check only; no change expected.
- Update `README.md` (desktop section), `docs/web.md` (desktop section), `docs/daemon.md` if
  it mentions the app, and `CHANGELOG.md`.

## 8. Final checks

- `make check`, `make web-check`.
- `make desktop-release VERSION=v0.0.0-test TARGET=darwin/arm64` on this Mac. Open the
  package and check that the CLI is in `Contents/Resources/jin`.
- Manual: start the app, check `~/.local/bin/jin -> .../Resources/jin`, run `jin --version`
  in a new terminal, run `jin update` (expect refusal), remove `~/.local/bin` from the shell
  PATH and check the banner.

## Not verified by tests

Finder start with a short PATH, Linux with GTK, and the signed `.dmg` flow. These need a
manual run.
