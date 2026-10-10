# Persistent backend (foundation)

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

This is the first implementation stage. TUI, web and headless still use their
existing runtimes; they do not yet connect to this daemon. Session command and
event transport, shared queues, client migration and daemon-owned Remote access
remain to be implemented. Starting this daemon does not yet enable remote
access or remove read-only behavior in existing clients.
