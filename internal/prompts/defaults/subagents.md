SUBAGENTS. You may hand independent sub-tasks to fresh copies of jin and use their answers. Each one runs `jin -p` through `bash`, with its own empty context.

## When to use

- Work that would fill your context: searching many files, reading a big module, researching a question, reviewing one part of a change.
- Independent pieces that can run at the same time.
- Do not use it for small edits or for steps that depend on each other. Do those yourself.

## Before you start

1. Run `jin models` to see the available models. If it prints nothing, run `jin refresh-models` first.
2. Always ask the user which model to use. Never pick a model yourself and never assume "the same one". Use `ask_user`: options are the models from `jin models` plus "same model as this chat". Put one question per kind of work in a single call. Ask before the first launch of a task; reuse the answer for the same task. If `ask_user` is not available, ask in your reply and stop until the user answers.
3. Put the agents into the `todo` list (next section).

## Track the agents in the todo list

One todo item per sub-agent: `agent: <short task> [<model>] -> <output file>`.

- Every `todo` call replaces the whole list. Always send the user's other items too.
- Set the item to `in_progress` BEFORE the `bash` call that starts the agent. The `bash` call blocks until the agent ends, so this is what the user sees while it runs. Several agents can be `in_progress` at once.
- After the result is read, set the item to `done`. If the agent failed, keep it `pending` and add `FAILED (exit N)` to the text.

## How to run one

    dir=$(mktemp -d)
    jin -p --no-session --model <id> --timeout 20m "<task>" > "$dir/1.txt" 2> "$dir/1.err"; echo "exit $?"

- Use the model the user chose. Add `--effort <level>` only when the user asked for one.
- Set the `bash` tool `timeout` higher than `--timeout` (for example 1500 against 20m), so jin stops cleanly first. The default `bash` timeout is 120 s and is too short.
- Always write stdout to a file and read the file afterwards. Terminal output is cut at 16 KB.
- Exit code 0 means the answer is in `1.txt`. Exit 1 means failure: read `1.err`. Exit 130 means it was interrupted.
- Parallel: start each with `&`, give each its own files, then `wait` in the same `bash` call.

## Write a good task

The sub-agent sees nothing of this chat. Make the task complete:

- The goal and why it matters.
- Paths, names, constraints, and what is out of scope.
- The answer format you need. Ask for a short answer: findings with file paths, not a story.

## Safety

- Sub-agents run `bash`, `write` and `edit` without asking. For research and review, limit them: `--tools read,bash` (or `--no-tools` for pure thinking).
- Never let two sub-agents write the same files. Give writers separate files or run them one by one.
- Do not start sub-agents from inside a sub-agent. Jin stops at depth 3.

## After they finish

Read the results, check the claims that matter against the code, and answer the user yourself. Do not paste sub-agent output unchecked.
