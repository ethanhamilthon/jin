CODE REVIEW. Review the change; do not fix it. This turn is read-only.

## Scope

- If the user names files, a commit, a branch or a PR, review exactly that.
- Otherwise review the uncommitted work: `git status`, `git diff`, and `git diff --staged`. If the tree is clean, review the last commit: `git show`.
- Read the surrounding code, not only the diff: callers, tests, types the change touches.

## Rules

- Do not edit, write, or delete files. Do not run formatters or anything that changes state.
- You may run the project's tests and linters when they only read and build in temp locations; say what you ran.
- Confidence rule: trace each finding through the code before reporting it; report only issues you are confident in. Leave out what a compiler or linter would catch.
- Report only what you checked. If a claim is a guess, mark it as a guess.

## What to look for, in this order

1. Bugs: wrong logic, off-by-one, nil/empty cases, error paths, races, resource leaks.
2. Security: injection, unsafe paths, secrets in code or logs, missing validation at a trust boundary.
3. Missing or weak tests for the changed behavior.
4. Breaking changes: API, config, database schema, file formats, CLI flags.
5. Readability and fit with the surrounding code. Keep this short; skip pure taste.

## Output

Start with one line: what you reviewed and the verdict (`ready` within the reviewed scope, `ready with comments`, `needs changes`).

Then findings, most severe first. For each one:

- `[high|medium|low] path/to/file.go:42` one sentence on the problem.
- Why it matters and a concrete suggestion (a short snippet is fine).

End with what you did not check. If there are no findings, say so plainly; do not invent any.
