# Context transparency (v0.13.0)

Date: 2026-10-09. Status: shipped in v0.13.0. Part of the v1 roadmap (section 11). Decisions come from the owner; what
is not decided is marked **proposed**.

## Goal

The owner and the user can see 100% of what the model receives. The context is the system
prompt and the hooks, and nothing else. Code adds no text of its own, with three named
exceptions that stay visible: tool schemas (the API needs them), the framing of messages
(section 5) and the summary that replaces a compacted history.

## What is decided

- Everything the system prompt holds is text in `system-prompt.md`, filled by `{{commands}}`:
  the docs pointer, the hooks, `AGENTS.md`, the environment and the session id.
- Only the `AGENTS.md` of the working directory is taken. Parent folders and the global one
  are not; a user who wants them writes a command or a hook.
- The macOS `sed -i` rule and the `TODO.md` line are gone from the code (done in v0.13.0
  work); the gallery has a hook for the macOS rule. Nested `AGENTS.md` files are no longer
  appended to tool results (done); the gallery has a hook.
- Hooks come through `{{jin hooks render}}`. It runs their commands in parallel
  (at most 8, 10 seconds each), skips disabled hooks, and prints global hooks and the
  project hooks of a trusted folder. No migration: a custom `system-prompt.md` without the
  line gets no hooks.
- The default `system-prompt.md` goes from the most stable text to the most unstable, with
  the cache break marker `<<jin-cache-break>>` (already in `internal/provider`) before the
  live part.
- Ships together with the `grep` tool as v0.13.0. Windows stays v0.14, benchmarks v0.15.

## 1. The default `system-prompt.md`

Order, top to bottom (proposed layout, one blank line between parts):

1. The role and the rules (static text, as today).
2. `{{jin docs}}`: the pointer to the documentation (stable).
3. `{{jin hooks render}}`: the hooks. Stable unless a hook runs a changing command.
4. `{{cat AGENTS.md 2>/dev/null; true}}` under a line `AGENTS.md:`. Changes when the file does.
5. `<<jin-cache-break>>`
6. The environment: `{{pwd}}`, `{{uname -sm}}`, `{{date +%F}}`, then `$JIN_SESSION_ID`.

The prompt of compaction and handoff stay sections of the same file.

`BuildSystemPrompt` stops adding anything: the system prompt is the filled file. The
`Tools: none` line goes (the tool schemas already say which tools exist); **proposed**.

## 2. Commands jin provides

- `jin docs`: prints the pointer text. `jin docs --list` lists the pages, `jin docs <name>`
  prints one page. All of `docs/*.md` is embedded in the binary, so it works offline and
  always matches the installed version. **Proposed** shape; search comes later.
- `jin hooks render`: as decided above. It reads the disabled and trusted settings
  from the database like `jin hooks list`.
- The prompt commands find this `jin`: the folder of the running binary is put first in
  `PATH` for them. `JIN_SESSION_ID` and `JIN_DIR` are set through `startup.Input.Env`.

## 3. What `/context` shows

The filled file is kept as segments: literal text, or the output of one command (with the
command). `/context` in the TUI and the Context window of jin web list every segment with
its full text, not only sizes, and the text is exactly what went on the wire. Pruned tool
results and compaction stay visible as their own entries.

## 4. Messages and tool results

- Tool results: `[exit code: N]`, truncation notes and the note on pruned old results stay;
  they are the protocol between a tool and the model.
- Framing of user input stays: `<pasted-prompts>`, `<attached-files>`, picture labels,
  `<system-refreshed>`, `<files-undone>`, `<task-result ...>`, `<background-tasks>`.
- Every such string is defined in one file (`internal/core/wire.go`) and documented in
  `docs/how-it-works.md`, so a review sees all of them.

## 5. Guard

A test builds the agent in all three modes (TUI, web, headless) and checks that the system
prompt equals the filled file and that every message string that code adds is in
`wire.go`. A new string elsewhere fails the test.

## 6. Order of work

1. `jin hooks render`, `PATH` and the env variables for prompt commands.
2. `jin docs` with embedded pages.
3. The default `system-prompt.md` and `BuildSystemPrompt` that adds nothing; `Tools: none`
   removed.
4. Segments and the full-text `/context` and Context window.
5. `wire.go` and the guard test.
6. Docs, `CHANGELOG.md`, release notes, v0.13.0.

Each step is a commit; steps 1 and 2 need no change in what the model sees.

## Risks

- Cache: the live part must stay after the marker. A hook with a changing command breaks
  the cache from its position on; the gallery says so.
- Start time: more commands run at start; each has 10 seconds and 8 run at once.
- Windows: `uname` and `date` come from Git Bash (v0.14).
- Headless: `jin -p -c` must keep the prefix stable; the date is after the marker.
- Custom prompts: users who edited `system-prompt.md` keep their file as it is.
