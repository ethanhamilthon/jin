# Benchmarks

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

