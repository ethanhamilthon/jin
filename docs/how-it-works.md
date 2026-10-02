# How jin works

## Layout of the code

```
main.go               wiring
internal/core         agent loop, system prompt, compact, handoff
internal/provider     OpenAI-compatible streaming client
internal/tools        read, write, edit, bash
internal/store        SQLite sessions and settings
internal/hooks        hook files
internal/prompts      reusable prompt files
internal/ui           terminal interface (tcell)
```

UI, core and provider are separate layers. The UI talks to the agent through channels.

## The agent loop

1. You send a message. Jin appends it to the history.
2. The history is streamed to the model together with the tool schemas.
3. If the answer has tool calls, jin runs them, appends the results and calls the model
   again. This repeats until the model answers without tool calls.
4. Messages you type while the agent works are queued and added after the current
   tools finish.
5. `Ctrl+C` interrupts the running turn. Unfinished tool calls get an "interrupted"
   result so the history stays valid.

Every message is saved to the database as it happens.

## Tools

- `read`: read a file, optionally with `offset` (1-based line) and `limit`. Pictures
  (png, jpeg, gif, webp, bmp) are attached so the model can see them.
- `write`: create a file or overwrite it.
- `edit`: replace an exact text match in a file.
- `bash`: run a shell command in the working directory. Default timeout 120 s, output
  is truncated at 16 KB.

Pasting an image saves it under `~/.jin/.pasted` and types its path. Ask the agent to
read that path.

## System prompt

Built when a session starts, from `internal/core/system_prompt.md` (embedded at build
time). Placeholders: `{{dir}}`, `{{os}}`, `{{date}}`, `{{hooks}}`, `{{jin_docs}}`,
`{{cat AGENTS.md}}`. Order in the final text: tools and guidelines, environment, jin
docs pointer (if enabled), enabled hooks, then the `AGENTS.md` block.

Edits to hooks, `AGENTS.md` and the Jin docs setting apply to new sessions only.

## Context files

Every `AGENTS.md` that applies is included:

1. The global one: `~/.jin/AGENTS.md`.
2. Ones in parent directories, marked "not the current project".
3. The one in the working directory.

The start screen lists the files used. Manage them in Esc → Context → AGENTS.md files.

## Compact and handoff

- **Compact** (Esc → Commands): the model summarizes the session. The chat shows a
  divider and the model continues from the summary. The full history stays in the
  database. It also runs on its own at 80% of the model's context window (windows come
  from the OpenRouter and LiteLLM catalogues).
- **Handoff**: the model writes a brief and jin opens a new session with that brief in
  the input, ready to edit.

## Data directory

Release builds use `~/.jin`, source builds (`make build`) use `~/.jin-dev`.

```
~/.jin/jin.db        sessions, messages, settings (SQLite, WAL)
~/.jin/AGENTS.md     global context
~/.jin/prompts/      reusable prompts
~/.jin/hooks/        hooks
~/.jin/.pasted/      pasted images
```

Files are created with `0600` permissions. Several jin processes can run at once and
share the database. The provider API key is stored in it as plain text.

## Provider and models

Set in Esc → Settings → Provider (base URL and API key). Select model picks the model
and reasoning effort. Scope models limits which models appear in the picker and the
`Ctrl+M` rotation. Effort is remembered per model.

## Settings menu

Select model, Scope models, Provider, Sound, Jin docs, Editor.

- **Sound**: Toggle, When (always or on blur), Volume. macOS plays a system sound,
  other systems get the terminal bell.
- **Jin docs**: Toggle. When on, new sessions get a pointer to these docs in the system
  prompt, so the agent fetches them when you ask about jin. Off by default.
- **Editor**: nano, vim or hx, used for editing prompts, hooks, `AGENTS.md` and the input.
