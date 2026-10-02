# What is Jin

Jin is a coding agent that runs in your terminal. You type a request, the model reads
files, edits them and runs commands, and you watch it happen. It speaks to any
OpenAI-compatible API (OpenAI, OpenRouter, local servers and so on).

Main traits:

- One binary, no server, no account. State lives in `~/.jin`.
- Four tools: `read`, `write`, `edit`, `bash`.
- Several sessions at once, saved per directory in SQLite.
- Streaming output with Markdown rendering. Images can be shown to the model.
- Everything the model sees is plain text you can read and change.

## Philosophy

**Minimalism.** Jin ships everything basic and nothing more. There is no plugin system,
no web search, no MCP layer, no built-in task planner. A feature that most users do not
need every day does not belong in the core.

**Extend with a CLI, not with code.** If the agent needs another capability, it uses the
shell. It can download or write a small CLI tool, then you describe that tool in a hook
(see [prompts-and-hooks.md](prompts-and-hooks.md)). From then on every session knows the
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
