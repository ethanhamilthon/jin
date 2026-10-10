# Managed proxy implementation plan

Status: implemented and locally checked; real-account acceptance remains pending.
No commit, push, or release requested.
Design: `2026-10-10-managed-proxy-design.md`.

1. Verify pinned upstream login, management, model discovery, prefix routing, and disabling.
   Stop if strict source isolation requires an upstream fork.
2. Add backward-compatible provider Disabled and managed-source metadata in store.
3. Add verified, bounded binary installation under paths.Global("cliproxyapi").
4. Add private loopback configuration and separate client/management keys.
5. Add authenticated shared process supervision, client/request leases, and dead-client cleanup.
6. Add local sign-in/status/cancel/logout for the three supported subscription profiles.
7. Add source-qualified catalog and client resolution, preserving ordinary API providers.
8. Gate every HTTP attempt without interrupting an already started response.
9. Wire source resolution through TUI, session.Manager, and headless.
10. Add atomic provider/model selection and provider-qualified effort memories.
11. Add provider switches and subscription setup to TUI/web Settings.
12. Add independent compatible version update/rollback to Settings.
13. Coordinate shutdown with reset/swap and document the supported behavior.
14. Run tests, race checks, frontend checks/build, and platform builds.
15. Ask the owner to run real-account login/stream/update checks; do not claim them passed.

Each stage gets focused tests. Keep files near 100 lines; split UI, orchestration,
management/auth, installation, and transport. Existing v0.13.2 fixes stay intact.

## Results

- Official v8.0.23 checksum, source-prefixed mock routes, OAuth URL/pending/cancel
  contracts, and actual installation through the web settings passed.
- Fake executable tests cover shared clients/request leases, shutdown, child crash
  recovery, update waiting for active requests, failed replacement rollback, and auth preservation.
- Provider/catalog tests cover duplicate IDs, partial failures, scopes, concurrent switches,
  atomic provider/model persistence, and disabling between HTTP retries.
- make check, targeted race tests, frontend checks/build, Linux amd64/arm64 builds,
  macOS amd64 build, and git diff --check passed.
- Browser smoke checks covered first-run subscription setup, accessible Settings before
  login/model selection, installation, switches, and 390px layout without console errors.
- Real accounts, refresh, actual model responses, cross-version auth migrations, and Linux
  runtime OAuth/browser behavior still require owner acceptance. No credentials were imported.
