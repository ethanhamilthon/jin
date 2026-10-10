# Managed CLIProxyAPI and enabled providers

Status: approved in chat; implementation starts with upstream contract validation.
The pending v0.13.2 fixes are independent. No release or new feature version is approved.

## Decisions

- Download CLIProxyAPI as a versioned external binary, not a Go SDK dependency.
- Use `paths.Global("cliproxyapi")`: production ~/.jin, isolated source-build ~/.jin-dev.
- macOS/Linux amd64/arm64 only. Windows uses the Linux build through WSL2.
- Initially expose Claude, Codex, and Antigravity as distinct managed provider entries.
- Every Jin provider, including ordinary API providers, has an enable switch.
- All enabled providers contribute to one model picker. A choice is provider ID + model ID.
- Identical model IDs from different providers remain distinct. No cross-provider fallback.
- Disabling lets an existing HTTP request finish but blocks every later request and retry.
- Disabling does not delete credentials or move a session to another provider.
- One managed proxy serves all Jin processes using the same data directory.
- Stop it when the last client and active request exit; recover dead-client registrations.
- Subscription login happens on the computer running Jin. Phones manage connected providers.
- Updates are explicit and independent of Jin, within the verified compatibility contract.

## Upstream acceptance gate

Candidate: CLIProxyAPI v8.0.23. Verify actual tagged code and CLI behavior, not stale SDK docs.
Confirm account login/status/cancel/logout, account disabling, model discovery, and strict
source-prefix routing when Claude models exist in both Claude and Antigravity.
If route isolation cannot be guaranteed without upstream modifications, stop and discuss it.
Real-account tests require the owner's interactive sign-in; never reuse local auth caches
implicitly or print OAuth tokens. Technical availability does not establish provider permission.
Anthropic's documented restrictions on third-party subscription login remain a release risk.

## Architecture

- `internal/cliproxy`: downloads/checksums, private config/auth, process coordination,
  local management client, sign-in, and compatible update/rollback.
- `internal/sources`: resolves saved providers and builds a provider-qualified model catalog.
- `internal/provider`: existing wire protocols, with a per-HTTP-request availability gate.
- `internal/store`: additive provider metadata; existing entries default to enabled.
- UI: TUI settings and Svelte settings panes; no credential storage or proxy logic in UI.

Separate managed source metadata from the OpenAI/Responses/Anthropic transport kind.
Keep `provider.active` as the existing default, not an exclusive enabled provider.
Existing sessions already store provider + model. Preserve both; distinguish deleted,
disabled, signed-out, and unreachable providers without silently rebinding.
Keep per-provider scope; key new effort memories by provider + model with legacy fallback.
Explicit headless --provider/--model behavior remains available.

## Safety and lifecycle

Bind both management and model access to loopback and authenticate them with different
random keys. Keep credentials private and out of SQLite API views, logs, SSE, and prompts.
Generate only the needed upstream configuration; disable automatic management UI updates,
network discovery, unneeded plugins, and cross-source failover where supported.

Coordinate ownership with an exclusive lock and authenticated local control, not an
unchecked PID file. Track clients and requests; update blocks new acquisitions and waits
for existing streams to close. Ordinary API use must not install or start the proxy.
Never terminate an independently launched proxy. Stop managed processes before reset/swap;
refuse a move while another Jin client still uses the data directory.

## Installation and updates

Install only after an explicit settings action, from official pinned release assets.
Verify checksum, exact binary entry, bounded extraction, and executable version before use.
Stage a new version separately; retain the previous binary and compatible configuration.
Test readiness without silently sending paid model prompts. Do not roll OAuth tokens back
blindly: refresh may invalidate older tokens. Block incompatible config/auth migrations.
A failed update must preserve the working version or report an actionable recovery state.

## UI

Settings, Providers holds ordinary providers and the three subscription entries, switches,
connection state, and local sign-in/logout. Existing provider entry points reuse this UI.
CLIProxyAPI settings show installation state, current/tested available versions, update,
rollback, and sanitized failures. A phone cannot initiate local-only sign-in.
Model pickers show the provider name and retain distinct provider/model identities.

## Checks

Mock upstream HTTP and executable fixtures cover downloads, corrupt archives, discovery,
source isolation, disabling during streams/retries, identical IDs, migrations, concurrent
clients, dead owners, update/rollback, and shutdown. Retain TUI/web/headless parity tests.
Run make check, targeted race tests, frontend checks/build, and both Linux cross-builds.
Manual account, OAuth browser, and real-proxy tests remain explicitly unverified until run.
