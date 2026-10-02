SUBAGENTS. You may hand independent sub-tasks to fresh copies of jin and use their answers. Each one runs `jin -p` through `bash`, with its own empty context. The only goal is speed: pieces of work run in parallel while you work too.

## Is it worth it

Before you launch anything, compare the time with and without sub-agents.

- Worth it: two or more independent pieces, each big enough (roughly over a minute) to pay for the start-up and for writing a full task; or a big read/search that you can do while you work on something else.
- Not worth it: small edits, one piece only, steps that depend on each other, or work you would finish before an agent even starts.
- If sub-agents give no speedup, say so to the user in one plain sentence and do the work yourself. Do not use them just because the user typed `#subagents`.

## Models

Edit this list to choose the models for sub-agents. One line per field: `name: model-id`. You may rename fields or add your own, for example `cheap: <id>`. Fields are empty by default.

smart:
fast:

- `smart` is for hard work: design, review, tricky bugs. `fast` is for simple work: search, reading many files, summaries. For any other field, use the one whose name fits the work best.
- If the field you need has a model id, use it and do not ask the user.
- If the field you need is empty, ask the user which model to use with `ask_user`. Ask once per task, in a single call for all empty fields you need; options are the models from `jin models` plus "same model as this chat". Reuse the answer for the same task. Never pick a model yourself and never assume "the same one". If `ask_user` is not available, ask in your reply and stop until the user answers.
- Do not write the answer into this file.

## Before you start

1. Run `jin models` only when you must ask the user, or when a listed id may be wrong. If it prints nothing, run `jin refresh-models` first.
2. Pick the model from the list above (see Models).
3. Put the agents into the `todo` list (next section).

## Track the agents in the todo list

One todo item per sub-agent: `agent: <short task> [<model>] -> <output dir>/N.txt`.

- Every `todo` call replaces the whole list. Always send the user's other items too.
- Set the item to `in_progress` when you launch the agent. Several agents can be `in_progress` at once.
- After the result is read, set the item to `done`. If the agent failed, keep it `pending` and add `FAILED (exit N)` to the text.

## Launch: always in the background

Never block on a sub-agent. The launch call must return at once, so you can keep working.

    dir=$(mktemp -d); echo "$dir"
    ( jin -p --no-session --model <id> --timeout 20m "<task>" > "$dir/1.txt" 2> "$dir/1.err" < /dev/null & echo $! > "$dir/1.pid"; wait $!; echo $? > "$dir/1.exit" ) > /dev/null 2>&1 &

- Start all independent agents in the same `bash` call, each with its own number N and its own files. Never run an agent in the foreground and never `wait` in the launch call itself (the `wait` inside the parentheses is the one that writes `N.exit`).
- Keep every redirect: without them `bash` waits for the agent to close its output.
- Use the model from the Models list or the one the user chose. Add `--effort <level>` only when the user asked for one.
- `N.exit` appears only when the agent has ended. It holds the exit code: 0 means the answer is in `N.txt`, 1 means failure (read `N.err`), 130 means interrupted.
- Always read results from the files. Terminal output is cut at 16 KB.

## Work, then poll

1. After the launch, do your own part of the task first. Do not sit idle.
2. When you have nothing left to do, or every 10 to 20 seconds of your own work, check the agents with one short call:

        sleep 15; ls "$dir"/*.exit 2>/dev/null

   Never sleep longer than 20 seconds in one call, and never poll in a tight loop.
3. For each new `N.exit`, read `N.txt` (or `N.err` on failure) and update the todo item.
4. Repeat until all agents have ended. Then answer.

## Write a good task

The sub-agent sees nothing of this chat. Make the task complete:

- The goal and why it matters.
- Paths, names, constraints, and what is out of scope.
- The answer format you need. Ask for a short answer: findings with file paths, not a story.

## Safety

- Sub-agents run `bash`, `write` and `edit` without asking. For research and review, limit them: `--tools read,bash` (or `--no-tools` for pure thinking).
- Never let two sub-agents write the same files, and do not edit files an agent is writing. Give writers separate files or run them one by one.
- Do not start sub-agents from inside a sub-agent. Jin stops at depth 3.
- Agents keep running after the launch call ends. Before you give the final answer, every agent must have an `N.exit` file or be stopped with `kill "$(cat "$dir/N.pid")"`. Jin also stops leftovers when it exits.

## After they finish

Read the results, check the claims that matter against the code, and answer the user yourself. Do not paste sub-agent output unchecked.
