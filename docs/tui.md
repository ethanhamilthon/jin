# Moving around the TUI

The screen has the chat on top and the input at the bottom. Focus is always either the
input or a panel. `Esc` switches between them.

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
| `Esc` | open the panel |
| `#` | start a prompt name; `Tab` or `Enter` completes it |

`Ctrl+M` needs a terminal that tells it apart from `Enter` (kitty keyboard protocol:
kitty, Ghostty, WezTerm, foot, recent iTerm2 and Alacritty).

## Mouse

- Wheel scrolls the chat.
- Select text with the mouse, then `Ctrl+C` copies it.

## The Esc panel

`Esc` opens the panel, `Esc` again closes it and returns to the input. Tabs, switched
with `←` and `→` (wrapping around):

| Tab | Contents |
| --- | --- |
| Commands | New session, Interrupt, Compact, Handoff, Quit |
| Sessions | Sessions of this directory. Green dot: the open one. Blinking blue dot: answering. Blue dot: unread answer |
| Prompts | Reusable prompts: edit, add, delete |
| Context | `AGENTS.md` files and Hooks |
| Settings | Select model, Scope models, Provider, Sound, Jin docs, Editor |
| Input | Focus input, Clear, Copy, Paste, Edit in editor |

Panel keys:

| Key | Action |
| --- | --- |
| `↑` `↓` | move between items, wrapping around |
| `Enter` | select |
| `←` `→` | switch tabs; on a row with choices (Sound, Jin docs, Scope) change its value |
| `Esc` | close the panel |

Lists without action keys have search always on: type to filter. Lists with action
keys read plain letters as actions and `/` opens search:

- Prompts: `Enter` edit, `a` add, `d` delete, `e` editor, `/` search.
- Hooks: `Enter` edit, `a` add, `d` delete, `t` on/off, `e` editor, `/` search.
- AGENTS.md files: `Enter` edit, `a` create one in the current directory (only when
  there is none).

Input actions keep the draft except Clear. Paste inserts at the cursor.

## Common tasks

- First setup: `Esc` → Settings → Provider (URL and key) → Select model.
- New session: `Esc` → Commands → New session.
- Switch session: `Esc` → `→` to Sessions → pick one.
- Change model quickly: `Ctrl+M`, or `Esc` → Settings → Select model.
- Use a prompt: type `#name` in a message.
- Add a hook: `Esc` → Context → Hooks → `a`.
- Quit: `Esc` → Commands → Quit. Jin asks first if a request is still running.

## Folding

`Ctrl+O` cycles three modes, saved between runs:

1. Everything: messages, reasoning and tool calls.
2. No tool calls: messages and reasoning.
3. Messages only.

The last chat line says what the next `Ctrl+O` does. When something is hidden and the
agent works, it shows `⠋ working...` with the last action.
