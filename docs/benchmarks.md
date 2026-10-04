# Benchmarks

## Latest run: 20 Python tasks (2026-10-04)

Aider Polyglot, one run per agent and task, 80 runs total. All agents used
`gpt-6-luna` with reasoning effort `high`. Jin was v0.7.0; these results do not
measure the v0.7.2 prompt changes.

| Agent | Passed | Agent time, s | Requests | Input | Cached | Output | Reasoning | Cost, USD |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| codex | 12/20 (60%) | 49 | 4.4 | 47 953 | 39 910 | 1 548 | 752 | 0.00198 |
| pi | 10/20 (50%) | 52 | 5.5 | 14 417 | 7 859 | 1 604 | 817 | 0.00154 |
| opencode | 9/20 (45%) | 106 | 11.4 | 79 724 | 62 413 | 2 221 | 698 | 0.00347 |
| jin (v0.7.0) | 8/20 (40%) | 47 | 6.0 | 16 985 | 8 832 | 1 616 | 654 | 0.00171 |

Time, requests, tokens and cost are averages per run. Cached tokens are part of Input;
Reasoning is part of Output. Tokens come from API usage recorded by a logging proxy.
Agent time excludes image builds and agent installation. No agent errors were reported.

Codex passed six tasks jin missed; jin passed two tasks codex missed. Codex was the only
agent to pass `bowling` and `connect`. Jin used fewer input tokens and cost less per run,
but passed fewer tasks. Twenty tasks with one run each are not enough to establish a
stable ranking or explain which agent behavior caused the difference.

The report and raw data are in the local `tbench` project: `RESULTS-r5.md`, `jobs/r5-*`
and `proxy/log.jsonl`.

## Earlier run: 10 Python tasks (2026-10-03)

These are early results from a small run: Aider Polyglot,
10 Python tasks, one run for each agent and task (50 runs in all), all agents on the same
model (`gpt-6-luna`, effort `high`). Numbers are the average of one run.

| Agent | Pass | Agent time, s | Requests | Input | Cached | Output | Reasoning |
| --- | --- | --- | --- | --- | --- | --- | --- |
| pi | 0.60 | 60 | 6.3 | 15 706 | 9 370 | 1 593 | 770 |
| opencode | 0.60 | 101 | 11.0 | 72 617 | 57 293 | 2 334 | 683 |
| codex | 0.50 | 48 | 4.1 | 43 592 | 34 842 | 1 645 | 835 |
| omp | 0.50 | 97 | 11.2 | 89 729 | 76 595 | 2 625 | 1 068 |
| jin (v0.4) | 0.50 | 56 | 6.1 | 16 929 | 9 984 | 1 708 | 816 |

Pass is the share of tasks with reward 1.0. The gap between agents is one task in ten,
which is within noise, so read this as "comparable", not as a ranking. What stands out is
the cost: jin sent about 17k input tokens per run, against 44k to 90k for codex, opencode
and omp. Tokens come from the `usage` field of the API responses, taken through a logging
proxy, not from the agents' own logs. Reasoning is part of Output. Agent time is the run
phase only, without building the image or installing the agent.

Caveats: ten tasks and one run each is a small sample; 11 of the 50 runs did not reach the
model (install timeout, network) and were restarted, so their times may be a little low; the
jin build was a local v0.4 build, not a release. Setup, task list and raw data are in the
`tbench` project.

