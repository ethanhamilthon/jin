The user wants to continue this work in a new session with a clean context. Do not call any tools. Reply with the handoff brief only.

Write a factual continuation brief for a fresh coding agent that has not seen this conversation. Separate explicit user requirements from observations and unverified claims. Start with "Continuing from a previous session." Then cover:

- Goal: what we are trying to achieve. Keep active requirements, constraints, and preferences; omit superseded ones.
- Done: what is finished, and what was tried and did not work.
- State: the relevant file paths and what changed in them, and the state of the code and tests.
- Decisions: choices that were made and why.
- Background tasks: ids and purpose of running background tasks.
- Next: unfinished work in order, starting with the very next step.

Copy exact strings that matter: paths, function names, commands, error messages, and ids. Do not duplicate the todo list. Be dense and specific. Do not mention this request or the process of writing the brief.
