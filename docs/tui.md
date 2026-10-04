# Moving around the TUI

The screen has the chat on top and the input at the bottom. Everything is typed into the
input: messages, `/commands`, `#prompts` and `@files`. Lists open above the input and
`Esc` closes them. With nothing open `Esc` does nothing, so it never stops a request.

The two status lines at the very bottom sit on a colored bar (blue in the default theme): the session title and model,
then the directory and usage: input and output tokens, the context against the model's window,
`cache NN%` (cached input tokens of the last request over its input tokens; hidden when the
provider does not report cached tokens) and the cost. The input always starts with `❯`; activity shows on the
rules above and below it instead (see Activity below).

## First run

Until jin has a provider and a model, it shows a first-run screen instead of the chat:
the version, the JIN logo, a short guide, the three API kinds (OpenAI Responses, OpenAI
Chat Completions, Anthropic) and the slogan, all centered.

1. Choose a kind with `↑` `↓` (or `1`-`3`) and `Enter`. Give the provider a name, the
   base URL (for example `https://api.openai.com/v1`, or `https://api.anthropic.com` for
   Anthropic) and the API key. `Esc` goes back to the kinds.
2. Pick a model and a reasoning effort. The chat opens.
3. Type a message. `Ctrl+C` or `/stop` stops a running request.

More providers can be added later with `/provider` → `a`. On the first-run screen, `s`
switches to another data folder, for example one that `/reset` put aside.

A session keeps the provider it started with. `/model` and `Ctrl+M` list the models of that
provider, with its scope, not of the active one. If the provider of a session was deleted,
the session cannot send: jin shows a message, and `/provider` → `Enter` on a saved provider
moves the session to it. Old sessions that saved no provider use the active one.

Only one jin process works in a session at a time. If another live process (a TUI or
`jin -p --session`) owns a session, `/sessions` opens it read-only: you see the history and
the line `Read-only: in use by process <pid>`, and nothing you type is sent or saved. If a
process takes a session you already have open, your next message is refused and the session
turns read-only. If the owner is gone, the session opens normally and open tool calls get an
"interrupted" result.

## Input

| Key | Action |
| --- | --- |
| `Enter` | send |
| `Shift+Enter` or `Alt+Enter` | new line |
| `←` `→` | move the cursor |
| `↑` `↓` | move the cursor between lines (the input scrolls) |
| `Home` `End` | start and end of the input |
| `Backspace` `Delete` | delete |
| `Ctrl+V` | paste at the cursor (an image is saved and its path typed) |
| `Ctrl+C` | copy the selection if there is one, otherwise interrupt the request |
| `Ctrl+M` | next model from the scope |
| `Ctrl+O` | next folding mode |
| `Ctrl+T` | edit the session todo list in the editor (same as `/todo`) |
| `/` | start a command; `Tab` or `Enter` runs it |
| `#` | start a prompt name; `Tab` or `Enter` completes it |
| `@` | start a file path; `Tab` or `Enter` completes it |

`Ctrl+M` needs a terminal that tells it apart from `Enter` (kitty keyboard protocol:
kitty, Ghostty, WezTerm, foot, recent iTerm2 and Alacritty).

Shortcuts and panel letter keys work in any keyboard layout: with a Russian layout on,
`Ctrl+м` is `Ctrl+V` and `ф` in a panel is `a`. Jin uses the physical key when the terminal
reports it. When it does not, jin maps the letters of the Russian, Ukrainian,
Belarusian, Kazakh and Greek layouts to their QWERTY keys.

## Mouse

- Wheel scrolls the chat.
- Select text with the mouse, then `Ctrl+C` copies it.
- Click a link in an answer to open it in the browser. Only `http`, `https` and `mailto`
  links open. Links are also OSC 8 hyperlinks, so terminals that support them open them
  their own way (often `Cmd`+click). Without a mouse, `/links` lists the links of the
  last answer and `Enter` opens one.
- Headings show as full-width colored bands.
- A line break inside a paragraph stays a line break, as in GitHub comments, so a
  heading line followed by `8. item` is not merged into one long line.
- Tables are drawn as grids. A table wider than the screen wraps its widest cells; when
  even that does not fit, each row is listed as `header: value` lines instead.

## Slash commands

Type `/` and a list of commands opens above the input, each with an icon and a short
description. Typing narrows the list by prefix. `↑` `↓` move, `Tab` or `Enter` runs the
highlighted command, `Esc` closes the list and leaves the text as it is.

A command works anywhere in the text, not only at the start. The `/` must start the text
or follow whitespace, so `/usr/bin`, `and/or` and URLs are never read as commands. When a
command runs, its `/name` is cut out of the draft (with one neighbouring space) and the
rest of the draft stays.

| Command | Does |
| --- | --- |
| `/sessions` | list the sessions of this directory (green dot: the open one, blinking blue dot: answering, blue dot: unread answer) |
| `/prompts` | reusable prompts, each with the start of its text: `Enter` edit, `a` add, `d` delete, `e` editor, `/` search |
| `/hooks` | hooks, each with the start of its text: `Enter` edit, `a` add, `d` delete, `t` on/off, `e` editor |
| `/model` | select the model and reasoning effort |
| `/scope` | choose the models `Ctrl+M` cycles through |
| `/provider` | providers: `Enter` use, `a` add, `d` delete |
| `/motion` | speed of the input glow and logo shimmer: off, slow, normal, fast |
| `/theme` | color theme, previewed as you move through the list; `n` new custom theme, `e` edit, `r` reload |
| `/sound` | notification sound: Toggle, When, Volume |
| `/tools` | switch agent tools on and off |
| `/async-tasks` | running background tasks of this directory: `Enter` shows the end of the output, `s` stops a task, `r` refreshes. |
| `/system-prompt` | edit the system, compaction and handoff prompts in `~/.jin/system-prompt.md` (created from the defaults) |
| `/change-editor` | choose the external editor |
| `/context` | show what fills the context: system prompt parts, tool schemas, conversation, the 5 largest tool results, cache share; token counts are approximate (bytes / 4) |
| `/compact` | summarize the conversation to free context |
| `/handoff` | have the model write a brief and continue in a new session |
| `/stop` | interrupt the running request |
| `/rewind` | restart the conversation from one of your messages in a new session; it does not change files |
| `/undo` | restore edit and write changes of the last turn; changes made through bash are not covered |
| `/diff` | open the diff of the last turn's edit and write changes in the editor |
| `/new` | new session; the rest of the draft moves into it |
| `/quit` | quit; jin asks first if a request is still running |
| `/clear` | clear the whole draft |
| `/links` | list the links of the last answer, `Enter` opens one |
| `/copy` | copy the draft to the clipboard |
| `/edit` | edit the draft in the editor |
| `/todo` | edit the todo list in the editor |
| `/tui <command>` | run a full-screen program |
| `/bash` | shell input |
| `/reset` | move the whole data folder (`~/.jin`) to a folder you name and quit; the next start is a fresh install. Settings, providers, prompts, hooks, themes and sessions all go with it |
| `/swap-config` | use another jin data folder instead of `~/.jin` and quit; the current data moves to that folder's place, so the same swap brings it back |

Menu commands (`/sessions`, `/prompts`, `/hooks`) open one panel with three tabs,
Sessions, Prompts and Hooks. `←` and `→` switch between them. Every panel is 6 rows
high and scrolls. Opening a panel keeps the draft, and `Esc` returns to it.

Panel keys:

| Key | Action |
| --- | --- |
| `↑` `↓` | move between items, wrapping around |
| `Enter` | select |
| `←` `→` | switch tabs; on a row with choices (Sound, Tools, Scope) change its value |
| `Esc` | close the panel (clears active filter first in Sessions) |

Lists without action keys have search always on: type to filter. Lists with action
keys read plain letters as actions and `/` opens search:

- Sessions: type to filter sessions live by title and user messages; `Backspace` edits
  the filter; `Esc` clears the filter first and closes the panel second; header line shows the active filter.
- Prompts: `Enter` edit, `a` add, `d` delete, `t` on/off, `e` editor, `/` search. System
  prompts are marked `system`: they can be switched off, not edited or deleted.
- Hooks: `Enter` edit, `a` add, `d` delete, `t` on/off, `e` editor, `/` search. System
  hooks (`async`, `docs`) work the same way as prompts.
- AGENTS.md files: `Enter` edit, `a` create one in the current directory (only when
  there is none).
- Providers: `Enter` use, `a` add, `d` delete, `/` search.

### /tui

`/tui lazygit` runs a full-screen program on the real terminal, the same way the editor
is started. jin suspends its screen and comes back when the program exits. The arguments
run to the end of the line, so `/tui htop -d 5` passes `-d 5`. Text before the command
stays in the draft. The command runs in the working directory through `sh -c`. It needs
an argument: press `Enter` after typing it.

### /bash

`/bash` replaces the input with a shell line (`$ `). Type a command and press `Enter`: it
runs in the working directory and its output goes to the chat. Stay in the mode and run
more commands until you press `Esc`. `Ctrl+C` stops the command that is running. The chat
draft is not touched. The output stays on the screen only: it is never added to what the
model sees, and the model does not know the commands were run. Output over 16 KB is cut and a
command stops after 10 minutes.

## Files with @

Type `@` to attach a file. A list of files opens, and it stays open while the text after
`@` is a valid path: `@src/`, `@~/notes/`, `@../other/`. Directories come first and end
with `/`; `Tab` on a directory goes inside it. Hidden files show up only when the name
starts with a dot. Long paths are shortened in the middle and keep the file name.

A name with spaces is put in quotes: `@"my file.md"`. You can also type the quote yourself,
and the list keeps working inside it.

The `@` must start the text or follow whitespace, so an email address is not a file.

On send, every `@path` that is a real file is replaced in the request by its full path, and
an XML block is added at the end:

```
<attached-files>
<file path="/Users/me/project/notes.md"/>
</attached-files>
```

The model reads the file itself with its tools. A path that does not exist, or is a
directory, stays plain text and is not listed. The chat shows what you typed.

## Common tasks

- First setup: the first-run screen → kind, name, URL, key → model and effort.
- New session: `/new`.
- Switch session: `/sessions`.
- Change model quickly: `Ctrl+M`, or `/model`.
- Use a prompt: type `#name` in a message.
- Attach a file: type `@path`.
- Add a hook: `/hooks` → `a`.
- Quit: `/quit`. Jin asks first if a request is still running.

## What's new and updates

After an upgrade, the first intro shows a short "What's new in v0.x" section, once. A
fresh install skips it.

When a session starts, jin checks in the background whether a newer release exists (at
most every 6 hours; the answer is kept in the settings `update.latest` and
`update.checked`). If there is one, the intro says so. Run `jin update` in a shell to
install it.

Long text wraps between words everywhere: messages, answers, the input, the todo list
and `ask_user`. Only a word longer than a whole row is split.

## Themes

`/theme` lists 10 built-in themes: Jin Original (the default), Tokyo Night, Catppuccin
Mocha, Gruvbox Dark, Nord, Dracula, One Dark, Rosé Pine, Solarized Light and GitHub
Light. Moving through the list shows each theme at once; `Enter` keeps it, `Esc` goes
back. A theme changes colors only. The choice is saved in the setting `ui.theme`.

The blocks above the input each have their own background: lists such as `/theme`, the
`/` commands, `@file` and `#prompt` completion, the todo list and `ask_user`.

### Custom themes

Custom themes are JSON files in `~/.jin/themes/`; `/theme` lists them after the built-in
ones. In `/theme`, `n` copies the highlighted theme into a new file with every color and
opens it in your editor; `e` edits a custom theme; `r` reloads the files.

```json
{
  "name": "My Night",
  "base": "Tokyo Night",
  "colors": { "bg": "#101018", "accent": "#7AA2F7", "todo_panel": "#132020" }
}
```

`base` is the built-in theme the missing colors come from (Jin Original when empty).
Colors are `#RRGGBB`. Keys: `bg`, `fg`, `text`, `muted`, `argument`, `detail`, `dim`,
`border`, `raised`, `status`, `on_status`, `status_title`, `accent`, `green`, `amber`,
`red`, `purple`, `pink`, `teal`, and the block backgrounds `panel`, `todo_panel`,
`ask_panel`, `slash_panel`, `files_panel`, `mention_panel` (left out, they are a light
tint of `bg`). A file with an unknown key or a bad color is skipped and named in red in
`/theme`.

`/motion` sets how fast the input glow and the logo shimmer move, or turns them off
(setting `ui.motion`). With motion off the input rules still change color while the
agent works, they just do not move. Spinners always turn, since they show that work
goes on. All animation pauses while the terminal window is not focused.

Jin follows the terminal's color depth. On a 256-color terminal the gradients (input
glow, logo shimmer, tinted tool rows) become plain colors, so they do not flicker. With
`NO_COLOR` set, or on a terminal without colors, the status bar and your messages use
reverse video and bold instead of colored bands. Set `TCELL_TRUECOLOR=disable` to force
the 256-color look.

## Rewind and undo

`/rewind` lists the messages you typed in this session, newest first. Choosing one
starts a new session that holds the history up to that message, with its text in the
input, ready to edit and send. The original session stays as it was. Files are not
touched; use `/undo` for them.

`/undo` restores every file that `edit` or `write` changed in the agent's last turn that
changed files: edited files get their old content back, created files are deleted. A
file you changed after the agent wrote it is left as it is and listed. When some files
changed since, `/undo` first shows a preview with `restore` or `skip (changed since)` per
file: Enter confirms, Esc cancels. With nothing to skip it restores at once. Repeat `/undo` to
go further back. The agent learns which files were restored with your next message.
Changes made by `bash` commands are not tracked; the `/undo` message says so. If some files
cannot be restored, jin lists the restored and the failed files, tells the agent about the
restored ones, and keeps the failed ones: run `/undo` again to retry. Run `/stop` first if the
agent works.

`/diff` opens the full diff of the files `edit` and `write` changed in the last turn in your
editor, so you can see what `/undo` would revert. Bash changes are not tracked.
`/diff session` is not available yet.

## Folding

`Ctrl+O` cycles four modes, saved between runs:

1. Tool output: everything, plus what each tool call produced under it. For `bash` the
   last 5 lines of its output; for `edit` and `write` the changed lines (red removed,
   green added), the last 5 of them. A first line `… N more lines` tells what was cut.
2. Everything: messages, reasoning and tool calls.
3. No tool calls: messages and reasoning.
4. Messages only.

The last chat line says what the next `Ctrl+O` does. When something is hidden and the
agent works, it shows `⠋ working...` with the last action.

If a `bash` command is running and you send the agent a message, the command moves to the
background at once: the tool call gets a task id and your message follows it.

While a new session starts, its input is closed: `Loading prompts...`, with the blue
glow running along the input rules, until the commands in the system prompt file, hooks and prompts are done. `Ctrl+C` skips the
commands that still run. The Prompts section of the intro shows a spinner after the name of
each prompt that is not ready yet.

In the session list, the dot in front of a session shows its state: green for the one on
screen; for the others a blinking blue dot while the agent works, a blinking purple dot while
a background task runs (also for sessions that are not open), a steady blue dot for an
unread answer, in that order of priority.

## Activity

A long glow runs from left to right along the rules above and below the input, wrapping
around without disappearing. Its bright segment is thicker; idle rules stay thin. It is blue while
the agent works (or a session loads, or a `/bash` command runs) and purple while the agent
is idle but the session has background tasks. The agent has priority. With nothing running
the rules are plain. The `JIN` logo on the start screen shimmers in a moving gradient.

When a task ends, a purple `async task <id> done` block appears in the chat (the first
lines of the output) and the agent gets the result as a message that is not yours.

## Todo and questions

The `todo` tool keeps a list for the session. While it has unfinished items it is
pinned above the input (max 7 rows, it scrolls). When every item is done the pin goes
away and the final list is added to the chat. The list is saved with the session.

`Ctrl+T` (or `/todo`) opens the list in the editor, one item per line:

```
- [ ] pending
- [~] in progress
- [x] done
```

Delete a line to remove an item; an empty file clears the list. Blank lines and one-line
`<!-- comments -->` are skipped, `*` works as a bullet and `x` or `X` marks done. A line
that is not an item is an error: nothing is saved and the temp file `jin-todo-*.md` is
kept, its path is shown in the chat. Without a list, `Ctrl+T` only prints
"No todo list yet".

The model learns about the edit twice: your next message starts with the edited list
(not shown in the chat), and its next `todo` update is refused once, with the current
list, so a stale list cannot overwrite your edit.

The `ask_user` tool replaces the input with its questions (max 7 rows). The last row of
every question takes free text.

| Key | Action |
| --- | --- |
| `↑` `↓` | choose an option; the last row is free text |
| `1`-`9` | jump to an option |
| `Enter` | answer |
| `←` | previous question (on the free-text row only when it is empty, otherwise it moves the text cursor) |
| `→` | next question, only if it already has an answer |
| `Space` | tick an option, in a question that allows several |

You can go back and answer a question again: the new answer replaces the old one. The
agent gets the answers after the last question. A question that allows several choices
shows `[ ]` and `[x]` boxes; `Enter` sends every ticked option (the option under the
cursor if none is ticked). Several answers reach the model as one bullet per line:

```
Which parts? →
- tests
- docs
```

Both blocks are hidden while a panel is open.
