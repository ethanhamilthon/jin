# What is Jin

Jin is a coding agent that runs in your terminal. You type a request, the model reads
files, edits them and runs commands, and you watch it happen. It speaks to OpenAI-compatible and
Anthropic-compatible APIs (OpenAI, Anthropic, OpenRouter, local servers and so on), and you can
save several providers and switch between them.

Main traits:

- One binary, no server, no account. State lives in `~/.jin`.
- Six tools: `read`, `write`, `edit`, `bash`, `ask_user`, `todo`. Each can be switched off with `/tools`.
- Several sessions at once, saved per directory in SQLite.
- Headless mode for scripts and CI: `jin -p "prompt"`, plus `jin models` and
  `jin refresh-models` (see [headless.md](headless.md)).
- A todo list per session that you can edit with `Ctrl+T`.
- Slash commands anywhere in the input, `@file` mentions, a `/bash` shell line and `/tui` for
  full-screen programs.
- Streaming output with Markdown rendering. Images can be shown to the model.
- Everything the model sees is plain text you can read and change.

## Philosophy

**Minimalism.** Jin ships everything basic and nothing more. There is no plugin system,
no web search, no MCP layer, no sub-agent framework (a sub-agent is just `jin -p`
started through `bash`). A feature that most users do not
need every day does not belong in the core.

**Extend with a CLI, not with code.** If the agent needs another capability, it uses the
shell. It can download or write a small CLI tool, then you describe that tool in a hook
(see [extending.md](extending.md)). From then on every session knows the
tool exists and calls it through `bash`. Jin itself stays small and the tool is yours.

**Clarity and transparency.** Nothing is hidden:

- The system prompt is one markdown file plus your hooks and `AGENTS.md` files. The
  start screen lists which ones were used.
- Tools run with no confirmation dialogs and no hidden rewriting. What you see in the
  chat is what ran.
- All data is in one SQLite file you can query with `sqlite3`
  (see [database.md](database.md)).
- Settings are plain keys in that same database.

**Trade-off.** Tools act directly on your machine. Run jin only in projects where
that is acceptable.
