- [x] agent: redesign tasks dialog with a one-hour filter [claude-haiku-5-5]
- [x] agent: pretty model names in format.ts, Scope, ModelList [claude-haiku-5-5]
- [x] agent: distinct notification sounds per event [claude-haiku-5-5]
- [x] agent: move Context and tool output into the composer, drop Undo [claude-haiku-5-5]
- [x] agent: new SwitchDialog picker component [claude-haiku-5-5]
- [x] agent: docs for the sidebar removal and the picker [claude-haiku-5-5]
# Research: improvements compatible with Jin

- [x] Review Jin philosophy and extension boundaries.
- [x] Find and read primary sources for tools and methodologies.
- [x] Rank recommendations by benefit, complexity, and fit; document limitations.

## Findings

1. Evaluate explicit RTK CLI calls for noisy shell outputs; retain raw output access and verify exit-code handling.
2. Adapt Superpowers debugging, verification, and TDD guidance into optional reusable prompts, not a skill framework.
3. Use targeted context retrieval and short documentation indexes rather than eager repository ingestion.
4. Consider ast-grep as an optional structural-search CLI described by a short hook.
5. Preserve progress and decisions in files for long tasks, using existing compact and handoff capabilities.

Primary sources read: Jin about/extending/how-it-works docs; RTK README sections; ast-grep quick start; Superpowers debugging, verification, and TDD sections; Anthropic context engineering and long-running harness sections; OpenAI harness engineering repository-knowledge section.

No tools were installed or benchmarked. Savings and quality improvements remain hypotheses, not measured Jin results. Existing pruning is implemented in internal/core/prune.go; recommendations should complement it rather than duplicate it.

Research only. No product changes, installations, or commits.


## Current roadmap work

- [x] Refresh shipped roadmap statuses and record successful phone checks.
- [x] Publish v0.13.1 Context dialog fix; GitHub release and CI succeeded.
- [x] Cancel native Windows at the owner's request; remove the unreleased implementation,
  installer, Windows CI/packaging changes, specs, and generated VM artifacts.
- [x] Restore unrelated platform/test changes from before the Windows work.
- [x] Preserve the HTTP 507 upstream-buffer retry and fast-task publication race fixes.
- [x] Prepare the fixes as v0.13.2 (Unreleased); do not tag, push, or publish a release.
- [x] Add WSL2 setup documentation and move the next workstream to benchmarks.
- [x] Check the final cleanup with make check, targeted race tests, npm run check, and git diff --check.
- [x] Inspect the existing tbench harness, local Python datasets, and prior reports.
- [ ] Agree on benchmark task sets, models, repetitions, and budget before execution.

Existing baseline: datasets/pg20 and the published v0.7.2 comparisons. An additional
local RESULTS-r7.md compares v0.8.1; its methodology needs review before inclusion.
Docker, uv, and harbor are available, but the Docker daemon is not running.
Linux amd64 and arm64 cross-builds passed; WSL2 runtime validation is still unverified.
No benchmark runs or paid API requests were started.

The owner confirmed native Windows worked in a VM, then chose WSL2 instead.
No WSL2 checks have run. macOS and Linux remain the supported native platforms.
The original research above is preserved. No pending native Windows work remains.

## Managed CLIProxyAPI and all-provider model selection

The owner approved implementation of the chat plan. Native Windows stays out of scope.
All providers have enable switches. Requests use the selected provider/model, with no
cross-provider fallback. Disabling lets the current request finish and blocks new ones.
One proxy serves TUI/web/headless and stops after the last client and request exit.
OAuth sign-in is local to the computer; phones manage already connected providers.

- [x] Validate v8.0.23 tagged management/prefix/disable APIs and executable contract.
- [x] Record the approved design and implementation plan in specs.

Official release checksum verified. Isolated local mock tests passed prefixed model
separation, strict Claude/Codex route selection, and start/poll/cancel for all three
OAuth providers. No account signed in, no existing credentials read, no model API bill.
Keep request cloaking/system-prompt replacement disabled in the managed configuration.
- [x] Implement provider state, unified catalog, and exact provider/model selection.
- [x] Implement verified binary installation, local supervision, local auth controls, and update/rollback.
- [x] Wire settings in TUI/web, request gates, and headless parity.
- [x] Add mock/executable-fixture tests for routing, switches, streams/retries, migrations,
  concurrent clients, child crash recovery, update idle wait, rollback, and auth preservation.
- [x] Document setup, lifecycle, version boundaries, provider restrictions, and validation limits.
- [x] Browser-check fresh subscription onboarding, Settings, verified real v8.0.23 installation,
  enable switches, and a 390px layout in an isolated temporary HOME (no account sign-in).
- [x] Complete final make check, targeted race tests, frontend checks/build, and Linux
  amd64/arm64 + macOS amd64 cross-builds after the onboarding fix.
- [ ] Owner-only validation: real sign-in, token refresh, model streams, and cross-version auth format.

No Go libraries were added. The proxy remains a separately installed binary. All tests
and browser checks use fixtures or temporary data, never the owner's subscription caches.
The cancelled-startup test now checks static instructions, not whether a fast command
happens to finish before cancellation. This matches the existing rendering contract.
Temporary browser/backend/Vite sessions were closed. No real model prompts were sent.

The unpublished v0.13.2 fixes remain separate. No commit, push, or release is requested.
