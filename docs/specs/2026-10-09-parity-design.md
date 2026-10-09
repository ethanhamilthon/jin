# Parity across TUI, web and headless (v0.10)

Date: 2026-10-09. Part of `2026-10-09-v1-roadmap-design.md`, section 9. Status: draft, waiting
for the owner's approval. Nothing here is built yet.

## Goal

The agent behaves the same in the TUI, in jin web and in `jin -p`. Interface differences stay
(themes, panes, key chords). Agent differences are either closed or listed as intentional.

## Scope (decided)

- One shared factory builds the agent for all three modes. It covers agent construction only.
- The start-up render (`startup.Input`, `finishRender`, reload) stays in `internal/ui` and
  `internal/session`. The TUI does not move onto `session.Manager` in v0.10.
- Headless stays text only and one-shot. It gets a `jin session` command group.

## Where the agent is built today

| Mode | File | Call |
| --- | --- | --- |
| TUI | `internal/ui/app_sessions.go` `startSessionAt` | `tools.BuildDir`, `NewAgent(client, "", ...)` |
| Web | `internal/session/start.go` `start` | `tools.BuildDir`, `NewAgent(client, "", ...)` |
| Headless | `internal/headless/agent.go` `buildAgent` | `tools.BuildHeadless`, `NewAgent(client, system, ...)` |

## Differences found, and the verdict for each

| Difference | TUI | Web | Headless | Verdict |
| --- | --- | --- | --- | --- |
| Tool registry | `BuildDir(dir)` | `BuildDir(dir)` | `BuildHeadless`, dir empty | Intentional: the bash tool says a command past its timeout is killed. The factory takes a mode. |
| `SetWorkdir` | set | set | not set | Equivalent today: headless calls `os.Chdir` (`cwd.go`) and core falls back to `os.Getwd()` (`agentsmd_nested.go`). Close: the factory always gets the directory. |
| `SetBackground` end with run | false | false | true | Intentional (see the roadmap: tasks end with the headless process). |
| `SetRefresher` | set at render | set at render | missing | Close. Without it a compaction in `jin -p` never re-renders the system prompt, while the TUI and web do. |
| `SetSystemPrompt`, `SetSidePrompts` | after the render | after the render | at construction | Close in shape: one `Apply(out)` function that all three call. |
| `SetContextSize` | after open | after open | in `prepare` | Close: a field of the factory input. |
| `SetRequestGate` | none | none | budget gate | Intentional: budgets exist only in `jin -p`. |
| Provider client | `clientFor` | `clientFor` | `provider.NewClient` | Out of scope. `clientFor` is duplicated in `ui` and `session`; noted for later. |
| Hooks and prompts | `startup.Render` | `startup.Render` | `startup.Render` | Same call, three copies of the input builder. The parity test compares them. |

## Design

Built as described below, with these changes while building: `Apply` sets the system prompt
and side prompts only (callers set the refresher with `agentkit.Refresher`, which also replaced
the two copies of `systemRefresher`); `SetContextSize` stays at the three call sites, because
a session knows its usage only after `start`; the one-shot registry keeps `BuildHeadless` with
an empty tool directory, because headless does `os.Chdir`; no `Wiring()` accessor was needed,
the test uses `SystemPrompt()` and `ToolSchemaBytes()`.

New package `internal/agentkit`, small files.

```go
type Mode int

const (
	Interactive Mode = iota // TUI and web
	OneShot                 // jin -p and jin session
)

type Spec struct {
	Mode        Mode
	Client      *provider.Client
	Names       []string
	Todos       tools.TodoStore
	Dir         string
	Owner       string // session id, or "headless" without a session
	System      string // empty until the render is done
	ContextSize int
}

func New(spec Spec) *core.Agent
func Apply(agent *core.Agent, out startup.Output, in startup.Input)
```

- `New` picks `BuildDir` or `BuildHeadless` by mode, then calls `SetWorkdir`,
  `SetBackground(tasks.Shared(), owner, mode == OneShot)` and `SetContextSize`.
- `Apply` sets the system prompt, the side prompts and the refresher. TUI and web call it in
  `finishRender`; headless calls it right after its render. This closes the refresher gap.
- `SetRequestGate` stays in `internal/headless`, because only it has budgets.
- Callers keep their own `Mode` input; nothing in `core` changes.

## Parity test

`internal/agentkit` test builds the agent in both modes from the same config and compares:

- tool names, order and schemas (the bash schema may differ only by the headless text);
- the `startup.Input` that each mode builds from the same config and trust state: tool names,
  disabled hooks and prompts, project hooks flag;
- the system prompt, compact and handoff texts that `startup.Render` returns for that input;
- that the refresher is set after `Apply`.

Open point: background wiring and compaction settings are private fields of `core.Agent`. To
compare them the test needs a small exported read-only `Wiring()` snapshot on the agent. A
new exported API is a design choice, so it needs approval (the alternative is a test inside
`internal/core` that takes the factory output).

## Parity table

`docs/parity.md`: feature by mode (TUI, web, headless), every gap either closed or marked
intentional, with the reason. It starts from the table above and adds slash commands and
settings. Building that list is the first step of the implementation: read `docs/tui.md`,
`docs/web.md` and `docs/headless.md` and mark what each mode lacks.

## Headless session commands

Built as `jin sessions compact|handoff|rewind|undo|context|reload <id>`, under the existing
`jin sessions list|search` group (decided: `jin session` next to `jin sessions` is easy to
mistype). The code is in `internal/headless/sessions_*.go` and drives `session.Manager`, so
each action is the web page's action. Changes from the draft below: `handoff` prints the
brief only, because the new session of the web flow is not saved until a first message, so a
one-shot process has no id to give; `rewind` without `--to` lists the messages, numbered from
1; the actions use the provider and model saved with the session, not `JIN_BASE_URL`.

Draft:

```
jin sessions compact|handoff|rewind|undo|context|reload <id>
```

Each is one-shot: open the session, build the agent with `agentkit.New` (mode `OneShot`), do
the one operation, save, exit. A unique id prefix is enough, as in `jin export`. Existing
logic to reuse:

| Command | Reuse |
| --- | --- |
| `compact`, `handoff` | `core.RequestCompact`, `core.RequestHandoff` (as in `session/side.go`) |
| `undo` | the revert plan of `session/undo.go` and `tools/revert.go` |
| `rewind` | `session.RewindPoints`; needs a turn argument (`rewind <id> --to <n>`) |
| `context` | `session.Context` report, printed as text, `--format json` for scripts |
| `reload` | `agent.ReloadPrompts`; in a one-shot process it only checks and prints that prompts and hooks render without warnings |

Rules:

- A session that another live jin process runs is refused, as for `jin -p` today.
- Output goes to stdout, errors to stderr, exit codes as in `internal/headless/run.go`.
- Flags and output texts are settled at the start of this step, with the owner.
- Code goes in `internal/headless`, files of about 100 lines, docs in `docs/headless.md`.

## Prompt prefix stability

`startup.Render` runs on every `jin -p -c`, so `{{commands}}` output can change the system
prompt and invalidate the provider cache. Step 1: measure on a real session, two runs, compare
the cached token share. Step 2 only if it is a problem; the fix is chosen after the numbers.

Result (2026-10-09): two renders of the default system prompt for the same input are
identical, so `jin -p -c` keeps the prefix. The date and the session id sit after the cache
break. Only a user's own `{{command}}` with changing output can break it, which
`docs/prompts-and-hooks.md` already says. No fix needed. Measured by comparing renders, not
by token counts from a provider.

## Order of work

1. Measure prompt prefix stability. Report the numbers.
2. Create `internal/agentkit` and move the three call sites onto it. Behavior stays the same
   except the closed gaps (refresher, workdir).
3. Parity test.
4. `docs/parity.md` with the full feature table; close or mark every gap it finds.
5. `jin session` commands, one command per commit, with tests and docs.
6. `CHANGELOG.md` entry, `make check`, `npm run check`.

## Out of scope

Pictures in headless, a long-lived headless process, moving the TUI onto `session.Manager`,
merging the two `clientFor` copies.
