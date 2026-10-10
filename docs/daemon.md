# Persistent backend

`jin daemon start` starts a detached local backend. It stays alive after the
command exits. `jin daemon status` reports its version, process id, agent state
and running task count. `jin daemon stop` shuts it down; `--force` is required
when agents or background tasks are busy.

The daemon owns a session manager and holds a lock on its data folder. Each data
folder has an independent daemon. Source builds use `~/.jin-dev`; release builds
use `~/.jin`. The control socket is local, private to the current OS user and
protected with mode 0600. Windows requires WSL for this implementation.

Clients must match the daemon version before issuing stop commands. A version
mismatch never restarts the daemon or interrupts work automatically. There is
no login-time autostart. Startup failures are recorded in `daemon.log` inside
the data folder.

Bare `jin`, `jin web`, session action commands, and saved-session continuations
with `jin -p --session` or `-c` automatically connect to the daemon. TUI and web
share one live agent and queue per session. Stop pauses the waiting queue;
`/resume` in the TUI or Resume in the browser continues it. Closing a client does
not interrupt an agent. Rewind creates a new session without changing the source.

Commands carry unique ids, so repeating an id returns the original result within
the lifetime of the daemon. After reconnect, clients reload snapshots. This is
not a durable exactly-once protocol across daemon crashes. The queue itself
lives in memory only. If the daemon dies during a turn, the session says so when
it is opened again: the request and any queued messages are not sent again.
New standalone headless requests and `--no-session` still use an isolated
runtime. Per-run tool overrides and budgets are not yet supported on shared
headless continuations.

Remote access belongs to the daemon. Enable it in `/settings`, Remote access,
or in the web Settings pane. It remains available without local clients and is
restored when the daemon starts again. Tailscale must already be installed,
signed in, and configured for HTTPS on the computer and phone. Jin refuses to
replace an HTTPS 443 Serve entry that points somewhere else. An entry jin
created itself, including one left by a daemon that crashed, is adopted again.
Pairing codes work once and expire in five minutes; device access can be renamed
or revoked from either interface.

A handoff reaches every client, but only the client that asked for it switches
to the new session. Stop pauses the waiting queue for that session in every
client, and Resume releases it again.

There is no automatic startup after reboot or login. Run `jin`, `jin web`, or
`jin daemon start` to start the daemon again. Data-folder reset and swap require
stopping the daemon first so its database is not moved while open.

`jin update` stops the running daemon before it replaces the binary and starts a
daemon of the new version afterwards, so an old daemon never keeps answering with
a version the new client refuses. A busy daemon stops the update with an error;
`jin update --force` interrupts the running agents and tasks and continues.
