# Managed CLIProxyAPI and providers

Status: unreleased implementation. The published v0.13.1 does not include these controls.
Jin still works with ordinary OpenAI-compatible, Responses, and Anthropic API providers.
The optional proxy is an external binary, not a Go library linked into Jin.

## Enable several providers

Web Settings, Providers and TUI `/settings`, Providers list all saved providers.
Use the rectangular switch in web or Left/Right in the TUI to turn a provider on/off.
The model picker lists models from all enabled providers, with the provider name.
Identical model IDs from different providers remain distinct. Each request uses the
chosen provider/model; errors and quota exhaustion never select another provider.

Disabling does not erase credentials or move a session. An HTTP request already running
finishes; later requests, including retries and tool-loop continuations, are blocked.
The default provider is only the default for new sessions, not the only enabled source.
Model scopes and remembered reasoning efforts are qualified by provider.

## Install the optional proxy

Open web Settings, CLIProxyAPI or TUI `/settings`, CLIProxyAPI. Choose Install.
Jin downloads an official pinned release, verifies SHA-256, validates the archive and
executable version, and checks the strict model-prefix contract before activation.

Production files live under `~/.jin/cliproxyapi`; source builds use `~/.jin-dev/cliproxyapi`.
Versioned binaries are in `versions/<version>/cli-proxy-api`; `installed.json` records
current/previous versions. Configuration and authentication files are private. macOS
and Linux amd64/arm64 are supported; Windows uses Linux through WSL2.

Installation is explicit. Subscription setup installs the newest compatible CLIProxyAPI by itself;
Settings can choose another v8.0.x version later. Ordinary API usage never installs or starts the proxy.
The generated configuration binds to 127.0.0.1, separates model/management/control keys,
disables plugins/discovery and management-panel updates, and disables request cloaking
that replaces system prompts or forces another client's identity.

## Connect subscriptions

Choose Add a provider, Subscription, then Claude, Codex, or Antigravity.
In the provider row choose Sign in. The browser flow runs on the computer hosting Jin.
The TUI uses `s` on a subscription row to open its connection controls; first-run `i`
opens proxy installation. Cancel aborts
pending sign-in. Sign out removes that profile's saved account files after requests finish.
Phones can enable/disable and manage connected providers, but cannot initiate local sign-in.

OAuth tokens belong to CLIProxyAPI's private `auth` directory. Jin does not copy existing
Claude/Codex/Google credential caches implicitly. Tokens are not returned through Jin's
web API or inserted into prompts. The proxy owns token refresh. Failed/expired login
requires an explicit retry; proxy failure does not fall back to an ordinary API provider.

The three profiles use distinct model prefixes: `claude/`, `codex/`, and `antigravity/`.
This keeps, for example, Claude through Anthropic separate from Claude through Antigravity.
Deleting a provider entry is not the same as signing out; sign out first to remove auth.

## Updates and lifecycle

Use Check versions and explicitly select a compatible v8.0.x release (v8.0.23 or newer).
A matching version number is only a candidate: checksum, executable-version, and prefix
contract checks must pass. New major/minor contracts require a Jin adaptation.

An update stages and probes the replacement without real credentials, blocks new proxy
request leases, waits for current requests, then changes the running version. The previous
version remains available for rollback. The generated configuration is recreated; OAuth
files are retained, not blindly restored to older refresh tokens. Do not assume an upstream
auth-format migration is reversible. Failure before activation leaves the current version
running; a failed replacement attempts to restart the previous version.

One authenticated local supervisor serves Jin processes sharing a data folder. Open
connections track clients and request leases, so process crashes release registrations.
The proxy stops when the last client, operation, and request have finished. If its child
exits, the next operation can restart it. Reset/swap closes the current client; the shared
data-folder lock prevents moving files while another client or supervisor uses them.

## Headless

Use the provider ID printed by `jin models`:

```sh
jin -p --provider subscription-codex --model codex/<model-id> "request"
```

`jin models` combines multiple saved providers; `--provider <id>` retains a single-provider
view. Setup, switches, and version changes are made in TUI/web Settings, not headless.
Subscription usage remains subject to plan quotas. Dollar figures are catalogue-based API
cost estimates, not measurements of subscription charges.

## Provider terms and validation

Technical OAuth support is not a licence to use a subscription as a third-party API.
Anthropic explicitly restricts third-party Claude subscription login and routing:
https://code.claude.com/docs/en/legal-and-compliance#authentication-and-credential-use
Review each provider's current terms before offering or using this integration.

Local fake-server and executable-fixture tests validate routing, switches, shared lifecycle,
and update mechanics. Official v8.0.23 mock routing and OAuth URL/pending/cancellation
checks passed without account sign-in. Real subscription sign-in, token refresh, model
streams, and cross-version auth compatibility still require owner validation.
