# Async tasks

Background tasks let the agent start a command and keep working. It does not wait: when
the task ends, the result comes back as a message. The agent learns how to use them from
the system prompt, which jin extends with the `jin async` instructions when the agent has
the `bash` tool and a session id (it is given as `Your session id`).

```
jin async run "<bash command>" --session <session-id>   # prints the task id, returns at once
jin async run --stdin "<bash command>" --session <id>   # the same, with a stdin pipe for input
jin async check --id <task-id> [--limit <chars>]        # status, exit code, end of the output
jin async input --id <task-id> "<text>" [--no-newline]   # write to the stdin of the task
jin async stop --id <task-id>                           # SIGTERM to the group, SIGKILL after 3 s
```

`check` without `--limit` prints the whole output. With `--limit N` it prints the last N
characters. The hook asks the agent to use `--limit` whenever it can, to keep its context small.

A task reads its stdin from `/dev/null`, so a command that reads stdin (`cat`, `jin -p`) gets
end of file at once instead of waiting forever. Only a task started with `--stdin` gets an open
pipe that `jin async input` writes to; `input` on any other task fails and says so. Inputs to
one task are written one at a time, and each gives up after 5 s when the task does not read
its stdin.

## The daemon

`jin async run` starts `jin daemon` when none runs (one per user: lock `~/.jin/async.lock`,
socket `~/.jin/async.sock`, mode 0600). The daemon runs each task as `bash -c` in its own
process group, with the environment and directory of the caller (`JIN_DEPTH` too, so the
depth limit of sub-agents holds), and writes stdout and stderr to `~/.jin/async/<id>.log`.
While a task runs its log stays under 64 MB: above that, the first and the last 16 MB are
kept with a note between them. Logs of finished tasks are removed a week after their last
write; the logs of running tasks are never removed, however quiet the task is.
Tasks live in the database (`async_tasks`), so they outlive the TUI, the terminal and the
agent that started them. The daemon quits after 10 minutes with no task. A daemon of
another jin version is replaced as soon as it has no running task. When a daemon starts it
settles the tasks of an earlier one: a dead process is marked `failed`, a live leftover is
stopped.

## Results wake the agent

When a task ends the daemon records its status and queues an event for the session
(`async_events`) in one database transaction:

```
<async-task-result id="3f9a1c20" status="done" exit="0">
...last 8000 characters of the output...
</async-task-result>
```

`status` is `done`, `failed` or `stopped`. A closing tag inside the output is broken up, so
output cannot pose as a message from the user.

A TUI opened in the directory of the session polls the database every second and takes
the events. If the agent is working, the message joins the conversation after its current
tool calls. If the agent has finished, the message starts a new turn. If the session is not
open, it is opened in the background and the focus stays where it was. Events of another
directory wait until a TUI opens there. The chat shows a short purple summary, not a user
bubble; the agent sees the full message and is told by the hook that it is not from the user.
`#prompts` inside a result are not expanded.

## `bash` moves to the background

The `bash` tool never kills a command just because it is slow. Two things send a running
command to the background:

- Its timeout (default 120 s, or the `timeout` argument) is reached.
- You send the agent a message while the command runs. The command moves to the background
  at once (after a second, when the message was already waiting), the agent gets the task
  id as the answer of the tool call, and your message follows it. You do not have to wait
  for the command or press `Ctrl+C`.

The result of the call has the output so far (up to 16 KB) and ends with the task id and
how to use it: `jin async check --id <id> --limit 2000`, `jin async stop --id <id>`. When the
command ends, its result arrives like any other task result. The agent can stop the task
later; you can too, with `/async-tasks`.

Details: the command keeps its process and its log in `~/.jin/async/<id>.log`; the daemon
only watches it. The exit code is written to `<id>.exit` by jin when the process ends. If
jin was closed before that, the task ends with status `ended` and no exit code. A task of
this kind has no stdin you can write to. When the daemon cannot take the command (for
example it does not start), the command is killed as before and the result says so.
`jin -p` has no background: its `bash` still kills a command at the timeout, because nobody
would read the result.

## Messages between agents

`jin async run "echo <message>" --session <id>` is a task like any other, and its output is
the message. A sub-agent uses it to tell its parent that it is done or has a question; the
parent does not have to wait, it wakes up by itself. The parent puts its own session id into
the task text of the sub-agent. `jin -p` does not wait for background tasks.

## `/async-tasks`

Lists the running tasks of this directory.
`Enter` puts the end of the output into the chat (for you, not for the agent), `s` stops a
task after a confirmation, `r` refreshes. When you stop a task, the agent gets
`<async-task-result ... status="stopped">stopped manually by the user</async-task-result>`.

While tasks run and the agent is idle, a purple glow runs along the input rules. A
working agent always has priority, with its blue glow.

Async tasks need a unix system.
