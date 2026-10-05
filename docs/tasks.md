# Background tasks

Background tasks let the agent start a command and keep working. It does not wait: when
the task ends, the result comes back as a message. The agent drives them with the `task`
tool:

| action | what it does |
| --- | --- |
| `start` | runs `command` (in `dir` when given) and returns the task id at once; `stdin: true` gives the task a stdin pipe |
| `check` | status, exit code and the end of the output (`limit` characters) |
| `input` | writes `text` and a newline to the stdin of a task started with `stdin` |
| `stop` | SIGTERM to the process group, SIGKILL after 3 s |
| `list` | the tasks of the session |

A task reads its stdin from `/dev/null` unless it was started with `stdin`, so a command that
reads stdin (`cat`, `jin -p`) gets end of file at once instead of waiting forever.

## Owned by jin

The jin process that started a task watches it: there is no daemon and no database table.
Each task runs as `bash -c` in its own process group and writes stdout and stderr to
`~/.jin/tasks/<id>.log`. While it runs, the log stays under 64 MB: above that, the first and
the last 16 MB are kept with a note between them. Logs are removed a week after their last
write. When jin exits, its tasks are killed.

## Results wake the agent

When a task ends, its result goes to the session that started it:

```
<task-result id="3f9a1c20" status="done" exit="0">
...last 8000 characters of the output...
</task-result>
```

`status` is `done`, `failed` or `stopped`. A closing tag inside the output is broken up, so
output cannot pose as a message from the user. A task the agent stopped itself sends no
result.

If the agent is working, the message joins the conversation after its current tool calls.
If the agent has finished, the message starts a new turn. If the session is not open, it
is opened in the background and the focus stays where it was. A session that cannot take
the result yet (read-only, provider deleted, prompts reloading) keeps it until it can.
The chat shows a short purple summary, not a user bubble.

## Before the final answer

When the agent answers without tool calls while some of its tasks still run, jin shows the
answer and sends the agent one note that lists those tasks. The agent stops the ones it no
longer needs and replies in one short line, or says in one line that they keep running.
That line is shown as the end of the answer, and the note itself stays hidden. This
happens once per turn.

## `bash` moves to the background

The `bash` tool never kills a command just because it is slow. Two things turn a running
command into a task:

- Its timeout (default 120 s, or the `timeout` argument) is reached.
- You send the agent a message while the command runs. The command moves to the background
  at once (after a second, when the message was already waiting), the agent gets the task
  id as the answer of the tool call, and your message follows it.

The result of the call has the output so far (up to 16 KB) and the task id. When the command
ends, its result arrives like any other task result.

## `jin -p`

`jin -p` has the `task` tool too, but its tasks are killed when the run ends, and their
results are not delivered: the agent checks them with `check` before it finishes. The note
before the final answer says so. Its `bash` still kills a command at the timeout.

## /tasks

Lists the tasks of this jin. `Enter` reads the end of the output, `s` stops a task after
confirmation, and `r` refreshes. When you stop a task, the agent gets
`<task-result ... status="stopped">stopped manually by the user ...</task-result>`.

While tasks run and the agent is idle, a purple glow runs along the input rules. A
foreground agent request uses the theme's primary color and takes priority.
