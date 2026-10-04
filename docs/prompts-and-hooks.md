# Prompts, hooks and extending jin

## Prompts

Reusable markdown snippets you call from a message with `#name`.

- Location: `~/.jin/prompts/` (`~/.jin-dev/prompts/` for source builds).
- A folder is part of the name: `review/security.md` is `#review/security`.
- Names may use letters, digits, `-`, `_`, `.` and `/`. No segment may start with a dot.
- Manage them with `/prompts`: `Enter` edit, `a` add, `d` delete, `e` editor.
- On send, each `#name` that matches a file is added to the request inside
  `<pasted-prompts>` as `<prompt name="...">`. The chat shows only what you typed.

Or just create the file yourself:

```sh
mkdir -p ~/.jin/prompts/review
printf 'Review the diff for security problems. Be concrete.\n' > ~/.jin/prompts/review/security.md
```

### Commands in prompts

Any text you write for jin can run a shell command and use its output. Put the command in
double braces: `{{git status --short}}` is replaced by what the command prints.

```markdown
Branch: {{git branch --show-current}}
Changed files:
{{git status --short}}
```

- Commands run once, when a session starts, not each time you send a message. The text
  that the model sees is fixed for the whole session, so `git status` shows the state at
  the start.
- Where: the system prompt file (below), hooks and `#prompts`. Never in `AGENTS.md` and
  never in the output of a task: those texts can come from a repository you cloned, and a
  repository must not run code just because you opened jin in it. Their braces stay as
  they are.
- Commands run with `bash -c` in the working directory and at the same time, at most 8 at
  once. Each has 10 seconds. Output has no size limit; trailing newlines are cut off.
- A command that fails or times out leaves `[command failed: ...]` in its place, and one
  line in the chat says which one. The rest of the text still works.
- `\{{` is a literal `{{`. An empty `{{}}` and a `{{` that is never closed are left alone,
  so JSON and templates in a prompt are safe.
- While the commands run, the session is on screen but its input is closed (`Loading
  prompts...`), and the names of the prompts that are not ready show a spinner in the
  intro. `Ctrl+C` stops the commands that still run; their places read `[command
  cancelled]` and the input opens.
- `#name` autocomplete offers the prompts of the session you are in. A prompt you edit
  applies to new sessions only.

### System prompts

Jin ships three prompts inside the binary: `#plan`, `#review` and `#subagents`. They are
marked `system` in `/prompts`. You cannot edit or delete them, and a file with the same
name in `~/.jin/prompts/` is ignored. Every prompt, system or not, can be switched off with
`t` (setting `prompts.disabled`); a switched-off prompt is not expanded and not offered
after `#`. The intro lists the enabled ones in a `Prompts` section.

Jin 0.3 copied these three into `~/.jin/prompts/` as ordinary files. On the first start of
0.4 jin moves `plan.md`, `review.md` and `subagents.md` to `~/.jin/prompts.bak/` (a copy
that already exists there is kept) and records `migrated.0_4` in the settings.

- `#plan`: plan mode. The agent explores, asks questions with `ask_user` and writes the
  plan into the `todo` list as `pending` items. It creates and edits no files.
- `#review`: read-only code review. A verdict line, then findings as
  `[high|medium|low] path:line`.
- `#subagents`: the agent starts sub-agents with `jin async run "jin -p ..."`, so it does
  not wait for them. The goal is speed: when parallel agents would not make the task
  faster, the agent tells you so and works alone.
  - Models: the agent runs `jin models` (the models you use, with price and context
    window) and asks you with `ask_user` which one to use. There are no model fields to
    fill in any more.
  - Results: a sub-agent gets the parent's session id in its task and, when it is done,
    runs `jin async run "echo ..." --session <parent-id>`. That message wakes the parent.
    See [async.md](async.md).

## The system prompt, compaction and handoff

Three prompts shape jin itself: the system prompt, the prompt used by `/compact`, and the
prompt used by `/handoff`. Each has a built-in default. To change them, type
`/system-prompt`: jin creates `~/.jin/system-prompt.md` from the defaults and opens it in
your editor. The compaction and handoff prompts are changed in the same file, there is no
separate command for them.

```markdown
# system

You are jin, ...

# compact

The conversation has grown long and must be compacted. ...

# handoff

The user wants to continue this work in a new session. ...
```

- A section starts with the exact line `# system`, `# compact` or `# handoff`. Any other
  line, including other headings, belongs to the section above it. Text before the first
  section line is ignored.
- A section that is empty or missing uses the built-in default. Delete the file to go back
  to the defaults; the intro shows `System prompt: custom (...)` while the file exists.
- The system section holds your own text only. Jin adds the rest, in this order: the list of
  tools, the jin docs pointer, the `jin async` instructions (see above), your hooks and the
  `AGENTS.md` files. After them come the working directory, the OS, the date and the session
  id; jin adds these itself, the default text has none. If your system section has its own
  line starting with `Environment:`, jin does not add its block (the session id is still added).

- Commands in all three sections work as described above.
- Edits apply to new sessions only. The file is global.

If you change a default text in the file, later versions of jin will not change it for
you, because the file is yours. Delete it, or the section, to get the new default.

## Hooks

A hook is a markdown file whose text goes into the system prompt at the start of every
new session.

- Location: `~/.jin/hooks/` (`~/.jin-dev/hooks/`) for global hooks, and
  `.jin/hooks/` inside a repository for project hooks (see below).
- Names: letters, digits, `-`, `_`, `.`; no folders.
- Manage with `/hooks`: `Enter` edit, `a` add, `p` add to the project, `d` delete,
  `t` on/off (new hooks are on), `e` editor.
- Enabled hooks are added alphabetically, global ones first, as plain text, before the
  `AGENTS.md` block. Empty hooks add nothing. Disabled names are stored in the setting
  `hooks.disabled` (a project hook by its full file path).
- Edits apply to new sessions only.

### Project hooks

A repository can ship its own hooks in `.jin/hooks/*.md`, so a team shares its tools and
habits the way it shares `AGENTS.md`. Because a hook's `{{commands}}` run on your machine,
jin asks once per folder, when it opens there and finds project hooks: "Run the project
hooks of this folder?". Until you say yes they stay off; the answer is stored in the
setting `hooks.trust`. In `/hooks` they are listed with `project`; `t` on an untrusted
one asks for trust. The intro marks them `(project)`. `jin -p` uses them only in a
folder you trusted.

### Sharing hooks

```
jin hooks add <url|path> [--name n] [--project]
jin hooks list
```

`jin hooks add` copies a markdown file into `~/.jin/hooks/` (with `--project`, into
`./.jin/hooks/`). The name comes from the file name unless you pass `--name`. It never
overwrites a hook. It is a copy, not a subscription: read it before you use it.

Use `NN-name.md` file names (for example `10-style.md`) to control the order.

Hook vs `AGENTS.md`: a hook is switchable and may run commands, `AGENTS.md` is tied to a
directory (or global at `~/.jin/AGENTS.md`). Use hooks for tools and habits you want
everywhere; `AGENTS.md` for project rules.

## Adding a capability: CLI + hook

A full walk-through with the reasons behind it is in [extending.md](extending.md).

Jin has no web search, no browser, no database client. When the agent needs one:

1. Get a CLI. Install an existing one (`brew install ripgrep`, `pip install ...`) or let
   the agent write a small script.
2. Put it on `PATH` (for example `~/.local/bin`).
3. Add a hook that says what the tool is and how to call it.

Example hook `~/.jin/hooks/20-websearch.md`:

```markdown
Web search: run `websearch "<query>"` through bash. It prints the top 5 results as
title, URL and snippet. Use it when the answer may be newer than your knowledge. Fetch
a result page with `curl -fsSL <url>`.
```

Keep the description short: what the command does, its arguments, what it prints, when
to use it. The agent will call it through `bash` like any other command. If the tool
breaks, you fix one script, not jin.

The jin docs pointer is not a hook. Jin adds it to the system prompt itself, always, so
that the agent can answer questions about jin. Jin also adds the `jin async` instructions
itself, but only when the agent has the `bash` tool and a session id (a `jin -p --no-session`
run has none). Neither can be switched off, and neither is a file you can edit.
