# Hooks and prompts gallery

Curated ready-to-use hooks and prompts for jin. Save the markdown below directly into
`~/.jin/hooks/` (global), `.jin/hooks/` (project), or `~/.jin/prompts/` (prompts),
or use:

```sh
jin hooks add <path-to-file> [--project]
```

> **Prompt caching note:** Place volatile `{{commands}}` at the end of hooks and prompts. Jin prefixes session context for LLM prompt caching; static text at the top stays cached across turns, while dynamic outputs at the end change without invalidating earlier prefix cache blocks.

---

## 1. Git status summary (Hook)

Injects repository status into the system prompt at session start.

- **File:** `10-git-status.md`
- **Location:** `~/.jin/hooks/10-git-status.md` (or `.jin/hooks/10-git-status.md`)
- **Install:** save markdown below to `10-git-status.md`, then run `jin hooks add ./10-git-status.md`

```markdown
# Git status

Repository status at session start:

Branch: {{git branch --show-current}}
Status:
{{git status --short}}
```

*Note: `{{commands}}` are placed at the end so preceding prompt text remains cached.*

---

## 2. Repo file list (Hook)

Provides tracked file structure so the agent navigates the repository without extra `bash` calls.

- **File:** `20-repo-files.md`
- **Location:** `.jin/hooks/20-repo-files.md` (or `~/.jin/hooks/20-repo-files.md`)
- **Install:** save markdown below to `20-repo-files.md`, then run `jin hooks add --project ./20-repo-files.md`

```markdown
# Tracked repository files

Tracked files in the repository:
{{git ls-files}}
```

*Note: Dynamic file listing is placed at the end to keep the system prompt prefix static.*

---

## 3. Tests-first policy (Hook)

Instructs the agent to reproduce bugs and verify new features with failing tests before writing production code.

- **File:** `30-tests-first.md`
- **Location:** `~/.jin/hooks/30-tests-first.md` (or `.jin/hooks/30-tests-first.md`)
- **Install:** save markdown below to `30-tests-first.md`, then run `jin hooks add ./30-tests-first.md`

```markdown
# Tests-first policy

When fixing bugs or adding new features:
1. Write a failing test reproducing the bug or asserting the expected behavior before editing implementation code.
2. Run the test suite to confirm the failure.
3. Make the smallest necessary code change to pass the test.
4. Run all tests to verify the fix and prevent regressions.
```

*Note: Pure static text without commands; highly cache-friendly.*

---

## 4. Release checklist (Prompt)

A checklist prompt called with `#release` before publishing a tag or release.

- **File:** `release.md`
- **Location:** save to `~/.jin/prompts/release.md`
- **Usage:** type `#release` in any prompt to expand

```markdown
# Release checklist

Verify release readiness before tagging:
1. Ensure working directory is clean (`git status`).
2. Run full test suite with race detector: `go test -race ./...`.
3. Run vet and static analysis: `go vet ./...`.
4. Verify version string and documentation updates.
5. Check recent commits:

Recent commits:
{{git log -n 5 --oneline}}
```

*Note: `{{git log}}` is positioned at the bottom to maintain cache stability for the checklist.*

---

## 5. Commit message style (Hook)

Enforces Conventional Commits rules across agent commits.

- **File:** `40-commit-style.md`
- **Location:** `~/.jin/hooks/40-commit-style.md` (or `.jin/hooks/40-commit-style.md`)
- **Install:** save markdown below to `40-commit-style.md`, then run `jin hooks add ./40-commit-style.md`

```markdown
# Commit message style

Follow Conventional Commits: `<type>(<scope>): <imperative summary>`:
- Types: feat, fix, refactor, perf, docs, test, chore, build, ci, style, revert.
- Subject: imperative mood, lowercase, <= 50 chars, no trailing period.
- Body: only when necessary to explain non-obvious motivation or breaking changes (wrap at 72 chars).
- Never add AI attribution, co-authors, or filler lines.

Recent project commit messages:
{{git log -n 3 --pretty=format:"%s"}}
```

*Note: The dynamic commit sample `{{git log}}` is placed at the end so style rules stay in prefix cache.*
