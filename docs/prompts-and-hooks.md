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

### Default prompts

Jin ships three prompts as ordinary files in `~/.jin/prompts/` (`~/.jin-dev/prompts/`
for source builds). When the TUI starts, jin creates the ones that are missing and never
touches the ones that exist, even when empty or edited. Headless commands (`jin -p`,
`jin models`) do not create them. A file you delete comes back at the next TUI start; to
change a default, edit the file.

- `#plan`: plan mode. The agent explores, asks questions with `ask_user` and writes the
  plan into the `todo` list as `pending` items. It creates and edits no files.
- `#review`: read-only code review. A verdict line, then findings as
  `[high|medium|low] path:line`.
- `#subagents`: the agent runs `jin -p` in the background through `bash` as sub-agents,
  works on its own part, polls them every 10-20 s and tracks them in `todo`. The goal
  is speed: when parallel agents would not make the task faster, the agent tells you so
  and works alone.
  - Models: the file has fields `smart:` and `fast:`, empty by default. Write a model id
    after a name (`smart: gpt-6-sol`) or add your own field (`cheap: <id>`). The agent
    uses a filled field without asking, and asks you with `ask_user` only when the field
    it needs is empty. It never writes your answer into the file. List ids with
    `jin models`.
  - Leftovers: agents keep running after the launch call returns; jin stops any that
    are still alive when it exits.

## Hooks

A hook is a markdown file whose text goes into the system prompt at the start of every
new session.

- Location: `~/.jin/hooks/` (`~/.jin-dev/hooks/`). Global, not per project.
- Names: letters, digits, `-`, `_`, `.`; no folders.
- Manage with `/context` → Hooks: `Enter` edit, `a` add, `d` delete, `t` on/off
  (new hooks are on), `e` editor.
- Enabled hooks are added alphabetically, as plain text, before the `AGENTS.md` block.
  Empty hooks add nothing. Disabled names are stored in the setting `hooks.disabled`.
- Edits apply to new sessions only.

Use `NN-name.md` file names (for example `10-style.md`) to control the order.

Hook vs `AGENTS.md`: a hook is global and switchable, `AGENTS.md` is tied to a
directory (or global at `~/.jin/AGENTS.md`). Use hooks for tools and habits you want
everywhere; `AGENTS.md` for project rules.

## Adding a capability: CLI + hook

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

## Jin docs in context

On by default; switch it with `/docs`. New sessions then get a short instruction in the
system prompt: when the user asks about jin, fetch these docs from GitHub and answer from them.
