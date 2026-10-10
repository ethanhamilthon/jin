import { get, post } from "./api";
import { app, fail } from "./app.svelte";
import { applyAccent } from "./accent";
import { askTrust } from "./trust";
import { replay } from "./events";
import type { AppState, Snapshot } from "./types";

export async function loadState() {
  const state = await get<AppState>("/api/state");
  app.version = state.version;
  app.dir = state.dir;
  app.home = state.home;
  app.latest = state.latest;
  app.config = state.config;
  app.project ||= state.dir;
  applyAccent(state.config.accent);
  app.loaded = true;
}

// keep stores a snapshot and replays the events that arrived before it.
export function keep(snap: Snapshot) {
  app.sessions[snap.state.id] = { state: snap.state, entries: snap.entries, intro: snap.intro, seq: snap.seq };
  replay(snap.state.id);
}

// show puts a session into a pane (the focused one by default).
export function show(id: string, pane = app.chatIndex) {
  const old = app.panes[pane]?.session;
  app.panes[pane].kind = "chat";
  app.panes[pane].session = id;
  app.focused = pane;
  app.project = app.sessions[id]?.state.path ?? app.project;
  if (old && old !== id) release(old);
  post(`/api/sessions/${id}/focus`).catch(() => {});
  if (app.sessions[id]?.state.unread) post(`/api/sessions/${id}/seen`).catch(() => {});
  askTrust(app.sessions[id]?.state.path ?? "").catch(() => {});
}

// release closes a session nobody looks at, unless it still works.
function release(id: string) {
  if (app.visible(id)) return;
  post<{ closed: boolean }>(`/api/sessions/${id}/close`)
    .then((r) => {
      if (r.closed && !app.visible(id)) delete app.sessions[id];
    })
    .catch(() => {});
}

export async function newSession(path = app.project || app.dir, pane = app.chatIndex) {
  try {
    const snap = await post<Snapshot>("/api/sessions", { path });
    keep(snap);
    show(snap.state.id, pane);
    return snap.state.id;
  } catch (err) {
    fail(err);
  }
}

export async function openSession(id: string, pane = app.chatIndex) {
  const shown = app.panes.findIndex((p) => p.kind === "chat" && p.session === id);
  if (shown >= 0) {
    app.focused = shown;
    return;
  }
  try {
    keep(await post<Snapshot>(`/api/sessions/${id}/open`));
    show(id, pane);
  } catch (err) {
    fail(err);
  }
}

// switchSession shows a session in the chat pane in focus, even when another
// pane already has it.
export async function switchSession(id: string) {
  try {
    if (!app.sessions[id]) keep(await post<Snapshot>(`/api/sessions/${id}/open`));
    show(id, app.chatIndex);
  } catch (err) {
    fail(err);
  }
}

let pending: Promise<void> | null = null;
let again = false;

// resync reloads the state and every open session. Events that arrive
// meanwhile wait and are replayed after their snapshot. A call made during
// a run asks for one more run.
export function resync(): Promise<void> {
  if (pending) {
    again = true;
    return pending;
  }
  pending = (async () => {
    try {
      do {
        again = false;
        await reload();
      } while (again);
    } finally {
      pending = null;
    }
  })();
  return pending;
}

async function reload() {
  app.resyncing = true;
  try {
    await loadState();
    for (const id of Object.keys(app.sessions)) {
      try {
        keep(await get<Snapshot>(`/api/sessions/${id}`));
      } catch {
        keep(await post<Snapshot>(`/api/sessions/${id}/open`));
      }
    }
  } finally {
    app.resyncing = false;
    for (const id of Object.keys(app.sessions)) replay(id);
  }
}

export function split() {
  if (app.panes.length >= 4) return app.toast("Four panes at most");
  app.panes.push({ key: app.nextKey++, kind: "chat", session: "" });
  newSession(app.current?.state.path, app.panes.length - 1);
}

export function closePane(index = app.focused) {
  const pane = app.panes[index];
  if (!pane || app.panes.length === 1) return;
  if (pane.kind === "chat" && app.panes.filter((p) => p.kind === "chat").length === 1) return app.toast("Keep one chat open");
  app.panes.splice(index, 1);
  app.focused = Math.min(app.focused, app.panes.length - 1);
  if (pane.session) release(pane.session);
}

export async function act(id: string, name: string, body?: unknown) {
  try {
    await post(`/api/sessions/${id}/${name}`, body);
    return true;
  } catch (err) {
    fail(err);
    return false;
  }
}
