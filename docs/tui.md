# Moving around the TUI

The work view shows the focused session or split panes above one shared input and status
area. Type messages, the supported work-view slash commands, `#prompts` and `@files` in the
input. Lists open above it. `Esc` goes back to the panel that opened the current one and
closes the first one. With nothing open, `Esc` does nothing.

Jin starts in a new session for the directory where you launched it. Everything else happens
in the work view through slash commands: `/sessions`, `/projects`, `/provider`, `/model`,
`/theme`, `/settings`, `/tasks` open their own list or flow above the input, and
`Esc` steps back through it. There is no second screen: the work view, its panes, drafts and running
agents stay where they are.

The two status lines at the very bottom sit on a colored bar (blue in the default theme): the session title and model,
then the directory and usage: input and output tokens, the context against the model's window,
`cache NN%` (cached input tokens of the last request over its input tokens; hidden when the
provider does not report cached tokens) and the cost. The input always starts with `❯`; activity shows on the
rules above and below it instead (see Activity below).

## First run

Until jin has a provider and a model, it shows a first-run screen instead of the chat:
the version, the JIN logo, a short guide, the choices of the current step and the slogan,
all centered.

1. Choose API or Subscription with `↑` `↓` (or `1`-`2`) and `Enter`.
2. For an API, choose one of three kinds (OpenAI Responses, OpenAI Chat Completions,
   Anthropic) with `↑` `↓` (or `1`-`3`) and `Enter`. Give the provider a name, the base URL
   (for example `https://api.openai.com/v1`, or `https://api.anthropic.com` for Anthropic)
   and the API key. For a subscription, jin first installs the newest compatible CLIProxyAPI
   by itself, then you choose Claude, Codex or Antigravity, sign in in the browser, and the
   model list opens at once. `Esc` goes back one step.
3. Pick a model and a reasoning effort. The chat opens.
4. Type a message. `Ctrl+C` or `/stop` stops a running request.

Manage saved providers with `/provider`. Changing the default provider affects new sessions;
existing sessions keep their saved provider, and `/model` and `Ctrl+M` select a model for the
focused session's provider. If the provider of a session was deleted, pick a new default with
`/provider`: the cut-off sessions move to it.

On the first-run screen, `s` switches to another data folder, for example one that the
Reset or Swap config rows of `/settings` moved aside.

Only one jin process works in a session at a time. If another live process (a TUI or
`jin -p --session`) owns a session, opening it from `/sessions` or `/projects` shows it read-only: you see the history and
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
| `Ctrl+V` | paste at the cursor (an image becomes an `[image 01]` token) |
| `Ctrl+C` | copy the selection if there is one, otherwise interrupt the request |
| `Ctrl+M` | next model from the scope |
| `Ctrl+O` | next folding mode |
| `Tab` | next pane; when a completion list or command list is open, `Tab` completes its selection instead |
| `/` | start one of the supported work-view commands; `Enter` runs the selection, `Tab` completes it |
| `#` | start a prompt name; `Tab` or `Enter` turns it into a prompt token |
| `@` | start a file path; `Tab` or `Enter` completes it |
| leading `$` | run a shell command when submitted; remove `$` to return to message input |

`Ctrl+M` needs a terminal that tells it apart from `Enter` (kitty keyboard protocol:
kitty, Ghostty, WezTerm, foot, recent iTerm2 and Alacritty).

### Tokens

Some parts of the draft are tokens: one colored element that the cursor steps over and
`Backspace` deletes whole.

- `[pasted text N lines]`: a paste of 3 or more lines or over 300 characters. Shorter
  pastes are typed as text. The model gets the pasted text exactly as it is: `#name`,
  `/command`, `@path` and `{{command}}` inside it are not expanded.
- `[image 01]`, `[image 02]`: a picture pasted with `Ctrl+V`. It is saved under
  `~/.jin/.pasted`; the model gets `[image 01: /path/to/file.png]`. Images are numbered in
  order within the message.
- `#name`: a prompt accepted from the list with `Tab` or `Enter`. Only a prompt token is
  expanded on send; `#name` typed as plain text (the list closed or skipped) stays literal.
- `/name`: a command accepted from the list or typed in full and followed by a space.
  `Enter` runs it instead of sending: the token is cut from the draft and the rest stays.
  A command that takes arguments, such as `/tui`, takes the text after it to the end of
  the line.

The chat shows the labels; the model gets the content. `/edit` gives tokens as plain text:
the pasted text, the image path, `#name`. A `#name` that comes back from the editor is plain
text and is not expanded.

Shortcuts and panel letter keys work in any keyboard layout: with a Russian layout on,
`Ctrl+м` is `Ctrl+V` and `ф` in a panel is `a`. Jin uses the physical key when the terminal
reports it. When it does not, jin maps the letters of the Russian, Ukrainian,
Belarusian, Kazakh and Greek layouts to their QWERTY keys.

## Mouse

- Wheel scrolls the chat.
- Select text with the mouse, then `Ctrl+C` copies it.
- Click a link in an answer to open it in the browser. Only `http`, `https` and `mailto`
  links open. Links are also OSC 8 hyperlinks, so terminals that support them open them
  their own way (often `Cmd`+click).
- Headings show as full-width colored bands.
- A line break inside a paragraph stays a line break, as in GitHub comments, so a
  heading line followed by `8. item` is not merged into one long line.
- Tables are drawn as grids. A table wider than the screen wraps its widest cells; when
  even that does not fit, each row is listed as `header: value` lines instead.

## Slash commands

Type `/` and a list of commands opens above the input, each with an icon and a short
description. Typing narrows the list by prefix. `↑` `↓` move, and `Enter` runs the
highlighted command immediately. `Tab` inserts a command token to run later with `Enter`.
Commands that need arguments are completed instead of run: add the arguments, then press
`Enter`. `Esc` closes the list and leaves the text as it is.

A command works anywhere in the text, not only at the start. The `/` must start the text
or follow whitespace, so `/usr/bin`, `and/or` and URLs are never read as commands. When a
command runs, its `/name` is cut out of the draft (with one neighbouring space) and the
rest of the draft stays.

| Command | Does |
| --- | --- |
| `/sessions` | sessions of this directory: browse, resume, search |
| `/projects` | registered directories: `Enter` switches, `a` adds an existing folder |
| `/model` | choose the focused session's model and reasoning effort |
| `/provider` | providers: add, switch, delete, on/off |
| `/theme` | color theme |
| `/settings` | global settings: sound, data folder, editor, tools, scoped models, motion, prompts, hooks, system prompt |
| `/reload` | rebuild the focused session's prompt from current hooks, prompts and instructions |
| `/compact` | summarize the conversation to free context |
| `/context` | show context parts, tool schemas, conversation, large tool results and cache share; `/context full` prints the exact text of every part |
| `/handoff` | write a brief and continue in a fresh session |
| `/rewind` | restart from one of your messages in a new session; it does not change files |
| `/stop` | interrupt the focused session's request or shell command |
| `/new` | create a session in the focused project; move the remaining draft into it |
| `/clear` | clear the whole draft |
| `/edit` | edit the draft in the editor |
| `/tui <command>` | run an interactive full-screen program on the real terminal |
| `/vertical` | split the focused pane into left and right panes |
| `/horizontal` | split the focused pane into top and bottom panes |
| `/quit` | close the focused pane; on the last pane, use the safe application shutdown flow |
| `/qa` | quit Jin through the safe application shutdown flow |
| `/tasks` | background tasks: output and stop |

These are the complete work-view slash commands.

### Panels

`/sessions` opens the sessions of the current directory, and `/settings` one list of the
global settings, each row opening its own panel: prompts, hooks, the system prompt, the
tool switches, the data folder and the rest. `↑` `↓` move, `Enter` opens or edits the highlighted entry,
`Esc` goes back to the list a panel came from, and from the list to the work view. Typing `/` in the panel filters the
sessions by content; `a`, `d`, `t` and `e` add, delete, toggle and open the editor, exactly
as the panel's own hint line says.

`/projects` lists the registered directories by path (the home folder is `~`). `Enter` switches the work view
to that directory (resuming its last session, or starting one), `a` registers an existing
folder, `d` removes one from the list after a confirmation (its sessions stay in the
database, and the open project cannot be removed). The folder field completes paths as you
type: `↑` `↓` choose a directory, `Tab` puts it in the field, `Enter` confirms the path.
Sessions belong to a project: `/new` creates one in the focused project and `/sessions`
lists the sessions of the current directory. `/tasks` lists the background tasks of this
jin, `Enter` shows the output and `s` stops a task after confirmation.

### Split panes

`/vertical` creates a left/right split. `/horizontal` creates a top/bottom split.
Jin supports up to four panes and rejects a split that would make a pane too small. Every pane
shows its project path and session title, and its own `ask_user` question at its
bottom. The focused pane has a blue border; while its session works a green glow runs around it.
Another pane's border is dark grey, with a blue glow while its session works. A purple glow
runs around any pane that waits on background tasks. The focused pane's title sits on a band of
the primary color. Click a pane, press `Tab`, or use
`Alt+Left`, `Alt+Right`, `Alt+Up` or `Alt+Down` to focus it.
`Tab` cycles the panes in layout order, wrapping at the end. While a completion list, the
command list or a panel is open, `Tab` completes their selection instead. All panes
share one input and status area, which act on the focused session. Drafts and scroll positions
belong to their sessions, and agents keep running when their panes lose focus or close.

`/quit` closes only the focused pane when other panes remain. Its session and draft remain
available in `/sessions`, and a running agent is not stopped. Closing the only pane follows the
safe application shutdown flow. `/qa` always requests full application shutdown.

### /tui

`/tui lazygit` runs a full-screen program on the real terminal, the same way the editor
is started. jin suspends its screen and comes back when the program exits. The arguments
run to the end of the line, so `/tui htop -d 5` passes `-d 5`. Text before the command
stays in the draft. The command runs in the working directory through `sh -c`. It needs
an argument: press `Enter` after typing it. Typing `/tui` and a space turns it into a
token; plain text such as a pasted `/tui x` does not run.

### Shell commands with $

A draft whose first character is `$` is a shell command. Type `$git status` or `$ git status`
and press `Enter`; Jin removes exactly the first `$`, runs the rest in the focused session's
project directory, and puts the command and output in that session's timeline. A bare `$` or
whitespace-only command does not run. More `$` characters, quotes and command text remain
unchanged. Removing the leading `$` returns the draft to normal message input.

Jin slash-command, prompt and file completion stay off while the draft begins with `$`.
`Ctrl+C` cancels the running command, and focus changes do not reroute its output. The output
is not automatically added to model context. The draft clears after a successful submission;
output is capped at 16 KB and the command times out after 10 minutes.

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

- First setup: the first-run screen → API or subscription → kind, name, URL, key (or sign in) → model and effort.
- New session in the current project: `/new`.
- Register a project: `/projects`, then `a`, and give an existing folder.
- Switch project: `/projects` and `Enter`; switch or resume a session: `/sessions` and `Enter`.
- Change model quickly: `Ctrl+M` or `/model`.
- Use a prompt: type `#` and pick the name from the list; manage prompts in `/settings`.
- Attach a file: type `@path`.
- Add a hook: `/settings`, then Hooks.
- Quit a pane: `/quit`; quit the application: `/qa`.

## What's new and updates

After an upgrade, the first intro shows a short "What's new in v0.x" section, once. A
fresh install skips it.

When a session starts, jin checks in the background whether a newer release exists (at
most every 6 hours; the answer is kept in the settings `update.latest` and
`update.checked`). If there is one, the intro says so. Run `jin update` in a shell to
install it.

Long text wraps between words everywhere: messages, answers, the input
and `ask_user`. Only a word longer than a whole row is split.

## Themes

`/theme` lists the built-in themes: Jin Original (the default),
Tokyo Night, Catppuccin Mocha, Gruvbox Dark, Nord, Dracula, One Dark, Rosé Pine, Solarized
Light and GitHub Light. Browsing previews a theme; `Enter` keeps it and `Esc` cancels. A
theme changes colors only. The choice is saved in `ui.theme`.

The blocks above the input each have their own background: the panel lists, `/` command
completion, `@file` and `#prompt` completion and `ask_user`.

### Custom themes

Custom themes are JSON files in `~/.jin/themes/`; `/theme` lists them after the built-in themes,
where `r` reloads the files.

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
`red`, `purple`, `pink`, `teal`, and the block backgrounds `panel`, `todo_panel`
(no longer used, still accepted so older theme files load), `ask_panel`, `slash_panel`, `files_panel`, `mention_panel` (left out, they are a light
tint of `bg`). A file with an unknown key or a bad color is skipped and named in red in
`/theme`.

The Motion row of `/settings` sets how fast the input glow and logo shimmer move, or
turns them off (`ui.motion`). With motion off, activity stays visible through color and
text. Spinners still turn to show ongoing work. All animation pauses while the terminal
window is not focused.

Jin follows the terminal's color depth. On a 256-color terminal the gradients (input
glow, logo shimmer, tinted tool rows) become plain colors, so they do not flicker. With
`NO_COLOR` set, or on a terminal without colors, the status bar and your messages use
reverse video and bold instead of colored bands. Set `TCELL_TRUECOLOR=disable` to force
the 256-color look.

## Rewind

`/rewind` lists the messages you typed in this session, newest first. Choosing one
starts a new session that holds the history up to that message, with its text in the
input, ready to edit and send. The original session stays as it was. Files are not
touched.

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

While a new session starts, its input is closed: `Loading prompts...`, with the theme's
primary-color glow running along the input rules, until prompt, hook and instruction commands
finish. `Ctrl+C` skips the
commands that still run. The Prompts section of the intro shows a spinner after the name of
each prompt that is not ready yet.

In `/sessions` and `/projects`, rows show model, activity and a text state such as
working, background task, unread, read-only or idle.

### Status dots

A dot shows the state of a session: green while the session is open in a work
pane, blinking blue while the agent answers outside the panes, blinking purple while a
background task of that session runs, and steady blue when the answer is unread. In
`/projects`, a `●` marks the current directory and `◐` one that has work in flight. Each pane
title shows its project path and session; a long path is cut from the left; the focused pane's title has the theme's primary color as
its background, like the status bar.

## Activity

A long glow runs from left to right along the rules above and below the input. Its bright
segment is thicker; idle rules stay thin. The foreground agent request uses the theme's
primary color. Background-task activity is purple. With no activity, the rules
are plain. The `JIN` logo on the start screen shimmers in a moving gradient.

When a task ends, a purple `background task <id> done` block appears in the chat (the first
lines of the output) and the agent gets the result as a message that is not yours.

## Questions

The `ask_user` tool shows its questions at the bottom of the session's pane (max 7 rows).
While that pane is focused, keys go to the questions instead of the input. The last row of
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
