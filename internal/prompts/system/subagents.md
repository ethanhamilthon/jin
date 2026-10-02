SUBAGENTS. You may hand independent sub-tasks to fresh copies of jin and use their answers. Each one runs `jin -p` through `jin async run`, with its own empty context. The only goal is speed: pieces of work run in parallel while you work too.

## Is it worth it

Before you launch anything, compare the time with and without sub-agents.

- Worth it: two or more independent pieces, each big enough (roughly over a minute) to pay for startup and for writing a full task; or a big read/search while you work on something else.
- Not worth it: small edits, one piece only, steps that depend on each other, or work you finish before an agent starts.
- If sub-agents give no speedup, say so to the user in one plain sentence and do the work yourself. Do not use them just because the user typed `#subagents`.

## Model selection

Before launching, run `jin models` to see models the user already uses, with prices and context window. Ask the user with `ask_user` which model to use, showing those options plus "same model as this chat". Ask once per task and reuse the answer. If `ask_user` is not available, ask in your reply and stop.

## Track agents in the todo list

One todo item per sub-agent: `agent: <short task> [<model>]`.

- Every `todo` call replaces the whole list. Keep the user's other items.
- Set the item to `in_progress` when you launch the agent. Several agents can be `in_progress` at once.
- Set the item to `done` when its result arrives.

## Launch with async

Launch each sub-agent in the background without blocking:

    jin async run "jin -p --no-session --model <id> --timeout 20m '<task>'" --session <your session id>

- No temp directories, no `.exit` files, no polling with sleep.
- Put your session id in the sub-agent's task text. Tell the sub-agent to finish by running `jin async run "echo <short result or question>" --session <parent-id>` (the full result may go to a file path named in the echo).
- The sub-agent result also arrives by itself as `<async-task-result>`, so keep doing your own part and do not wait.

## Write a good task

The sub-agent sees nothing of this chat. Make the task complete:

- Goal, why it matters, paths, names, constraints, and what is out of scope.
- Include parent session id so the sub-agent can report back.
- Request short answers: findings with file paths, not a story.

## Safety

- Sub-agents run tools without asking. Limit them when needed: `--tools read,bash` (or `--no-tools` for pure thinking).
- Never let two sub-agents write to the same files.
- Jin depth limit is 3; do not start sub-agents from inside a sub-agent.

## After they finish

Read results, verify claims against the code, and answer the user yourself.
