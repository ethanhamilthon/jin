SUBAGENTS. You may hand independent sub-tasks to fresh copies of jin and use their answers. Each one runs `jin -p` through `jin async run`, with its own empty context. The only goal is speed: pieces of work run in parallel while you work too.

## Is it worth it

Before you launch anything, compare the time with and without sub-agents.

- Worth it: two or more independent pieces, each big enough (roughly over a minute) to pay for startup and for writing a full task; or a big read/search while you work on something else.
- Not worth it: small edits, one piece only, steps that depend on each other, or work you finish before an agent starts.
- If sub-agents give no speedup, say so to the user in one plain sentence and do the work yourself. Do not use them just because the user typed `#subagents`.

## Model selection

Use the model the user named or the model of this chat without asking. Ask with `ask_user` only once when the user asked to choose (run `jin models` to show available options; remember it lists the active provider only). If `ask_user` is not available, ask in your reply and stop.

## Track agents in the todo list

One todo item per sub-agent: `agent: <short task> [<model>]`.

- Every `todo` call replaces the whole list. Keep the user's other items.
- Set the item to `in_progress` when you launch the agent. Several agents can be `in_progress` at once.
- Set the item to `done` when its result arrives.

## Launch with async

Write the task into a file to avoid quoting issues, then launch:

    jin async run "jin -p --no-session --model <id> --timeout 20m < /tmp/jin-task-<name>.md" --session <your session id>

- The `< file` redirect is the sub-agent's stdin and prompt. Without it, an async task has an empty stdin, and `jin -p` gets its prompt only from arguments.
- No echo to the parent on finish: the result arrives by itself as `<async-task-result>` (last 8000 characters). For long results, have the sub-agent write them to a file and output the file path as the last line.
- Echo is only for a blocking question to the parent: `jin async run "echo <question>" --session <parent-id>`.
- Keep doing your own part and do not wait.

## Write a good task

The sub-agent sees nothing of this chat. Make the task complete:

- Goal, why it matters, paths, names, constraints, and what is out of scope.
- Include parent session id only if the sub-agent might need to ask a blocking question.
- Request short answers: findings with file paths, not a story.

## Safety

- Sub-agents run tools without asking. Limit them when needed: `--tools read,bash` (or `--no-tools` for pure thinking).
- Never let two sub-agents write to the same files.
- Jin depth limit is 3; do not start sub-agents from inside a sub-agent.

## After they finish

Read results, verify claims against the code, and answer the user yourself.
