# Benchmarks

## grep tool and repo map: search questions on large repositories (2026-10-09)

Jin v0.12.0 plus the `grep` tool, model `gpt-6-luna` with reasoning effort `medium`, read-only
questions with one right answer (the file of an error message, the file of a route, a count of
files). Three setups: A without the `grep` tool, B with it, C with it and a repo map in the
prompt (the gallery recipe). Every answer was checked; all 72 runs were correct.

| Repository | Setup | Runs | Input tokens | Time | Tool calls |
| --- | --- | --- | --- | --- | --- |
| jin (935 files, not in the model's training data) | A | 12 | 4376 | 9.5 s | 1.2 |
| | B | 12 | 4833 | 4.6 s | 1.0 |
| | C | 12 | 5954 | 4.8 s | 1.1 |
| Kubernetes (31,378 files) | A | 12 | 4309 | 6.4 s | 1.0 |
| | B | 12 | 4760 | 7.8 s | 1.0 |
| | C | 12 | 9711 | 9.4 s | 1.2 |

The model chose the `grep` tool over `bash` in 21 of 24 runs when it had it. On the jin
repository it halved the time; on Kubernetes the model knows the code and often needed no
search, so the setups were equal. The repo map doubled the input tokens on Kubernetes and
helped nowhere, so it stays a gallery recipe and is not a default. Small samples and easy
questions: this measures the speed of finding a place, not the quality of changes.

## Claude Sonnet 5.5: 10 Python tasks (2026-10-04)

Aider Polyglot, 10 Python exercises (`dot-dsl`, `hangman`, `paasio`, `pov`, `react`, `rest-api`, `scale-generator`, `tree-building`, `zebra-puzzle`, `zipper`), one run per agent and task. Model: `claude-sonnet-5-5` with reasoning effort `medium`. Tested agents: claude-code, jin v0.7.2, pi, opencode.

Pricing: $3.00/M input, $0.30/M cached input, $15.00/M output.

| Agent | Passed | Agent time, s | Requests | Input | Cached | Output | Cost, USD |
| --- | --- | --- | --- | --- | --- | --- | --- |
| claude-code | 8/10 (80%) | 13 | 3.3 | 77 606 | 70 512 | 1 189 | 0.06028 |
| **jin (v0.7.2)** | 7/10 (70%) | 15 | 3.5 | 12 976 | 10 169 | 1 389 | 0.03231 |
| pi | 7/10 (70%) | 16 | 3.4 | 13 055 | 9 802 | 1 383 | 0.03344 |
| opencode | 7/10 (70%) | 39 | 5.0 | 50 230 | 41 991 | 1 867 | 0.06532 |

Per-task results:
- `hangman`, `react`, `rest-api`, `scale-generator`, `zebra-puzzle`, `zipper`: passed by all 4 agents.
- `tree-building`: passed by claude-code, jin, and pi; failed by opencode.
- `pov`: passed by claude-code and opencode; failed by jin and pi.
- `dot-dsl`, `paasio`: failed by all 4 agents.

Raw data and job configs: `tbench/RESULTS-c1.md`, `tbench/jobs/c1-*`.

## GPT-6 Luna: 20 Python tasks (2026-10-04)

Aider Polyglot, 20 Python exercises, one run per agent and task. Model: `gpt-6-luna` with reasoning effort `high`. Tested agents: codex, jin v0.7.2, pi, opencode, jin v0.7.0.

Pricing: $0.10/M input, $0.01/M cached input, $0.50/M output.

| Agent | Passed | Agent time, s | Requests | Input | Cached | Output | Reasoning | Cost, USD |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| codex | 12/20 (60%) | 49 | 4.4 | 47 953 | 39 910 | 1 548 | 752 | 0.00198 |
| **jin (v0.7.2)** | 11/20 (55%) | 65 | 7.7 | 24 880 | 14 848 | 2 323 | 993 | 0.00231 |
| pi | 10/20 (50%) | 52 | 5.5 | 14 417 | 7 859 | 1 604 | 817 | 0.00154 |
| opencode | 9/20 (45%) | 106 | 11.4 | 79 724 | 62 413 | 2 221 | 698 | 0.00347 |
| jin (v0.7.0) | 8/20 (40%) | 47 | 6.0 | 16 985 | 8 832 | 1 616 | 654 | 0.00171 |

Time, requests, tokens and cost are averages per run. Cached tokens are part of Input; Reasoning is part of Output. Tokens come from API usage recorded by the logging proxy. Agent time excludes image builds and agent installation.

Raw data and job configs: `tbench/RESULTS-r5.md`, `tbench/RESULTS-r6.md`, `tbench/jobs/r5-*`, `tbench/jobs/r6-*`.

## Earlier run: 10 Python tasks on GPT-6 Luna (2026-10-03)

Initial run on 10 tasks (`affine-cipher`, `bowling`, `connect`, `dominoes`, `food-chain`, `grep`, `pig-latin`, `poker`, `transpose`, `wordy`). Model: `gpt-6-luna`, effort `high`.

| Agent | Pass | Agent time, s | Requests | Input | Cached | Output | Reasoning |
| --- | --- | --- | --- | --- | --- | --- | --- |
| pi | 0.60 | 60 | 6.3 | 15 706 | 9 370 | 1 593 | 770 |
| opencode | 0.60 | 101 | 11.0 | 72 617 | 57 293 | 2 334 | 683 |
| codex | 0.50 | 48 | 4.1 | 43 592 | 34 842 | 1 645 | 835 |
| omp | 0.50 | 97 | 11.2 | 89 729 | 76 595 | 2 625 | 1 068 |
| jin (v0.4) | 0.50 | 56 | 6.1 | 16 929 | 9 984 | 1 708 | 816 |
