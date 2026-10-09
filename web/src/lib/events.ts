import { app } from "./app.svelte";
import { keep, loadState, resync, show } from "./actions";
import { ring } from "./sound";
import { get, post } from "./api";
import type { ServerEvent, Snapshot } from "./types";

export function connect() {
  let first = true;
  const source = new EventSource("/api/events");
  source.onopen = () => {
    app.connected = true;
    if (!first) resync().catch(() => {});
    first = false;
  };
  source.onerror = () => {
    app.connected = false;
  };
  source.onmessage = (message) => handle(JSON.parse(message.data) as ServerEvent);
  return new Promise<void>((resolve) => source.addEventListener("open", () => resolve(), { once: true }));
}

const early = new Map<string, ServerEvent[]>();
let replaying = false;

// replay applies the events of a session that came before its snapshot.
export function replay(id: string) {
  const events = early.get(id) ?? [];
  early.delete(id);
  replaying = true;
  try {
    for (const ev of events) handle(ev);
  } finally {
    replaying = false;
  }
}

function handle(ev: ServerEvent) {
  const view = ev.session ? app.sessions[ev.session] : undefined;
  const held = app.resyncing && !replaying && ev.session && view;
  if (ev.session && (!view || held) && ev.type !== "ring" && ev.type !== "handoff") {
    const list = early.get(ev.session) ?? [];
    if (list.length < 5000) list.push(ev);
    early.set(ev.session, list);
    return;
  }
  if (view && ev.seq <= view.seq) return;
  if (view) view.seq = ev.seq;
  switch (ev.type) {
    case "entry":
    case "update":
      if (view && ev.entry) view.entries[ev.index ?? 0] = ev.entry;
      break;
    case "delta":
      if (view) view.entries[ev.index ?? 0].text += ev.text ?? "";
      break;
    case "entries":
      if (view) view.entries = ev.entries ?? [];
      break;
    case "state":
      if (view && ev.state) view.state = ev.state;
      break;
    case "ring":
      onRing(ev);
      break;
    case "handoff":
      onHandoff(ev);
      break;
    case "sessions":
      app.sessionsRev++;
      break;
    case "projects":
      app.projectsRev++;
      break;
    case "tasks":
      app.tasksRev++;
      break;
    case "config":
      app.configRev++;
      loadState().catch(() => {});
      break;
    case "notice":
      app.toast(ev.text ?? "");
      break;
    case "resync":
      resync().catch(() => {});
      break;
    case "quit":
      app.stopped = true;
      break;
  }
}

function onRing(ev: ServerEvent) {
  const id = ev.session ?? "";
  const looking = app.pane.session === id;
  if (ev.kind === "ask" || ev.text === "final") ring(app.config.sound, document.hasFocus());
  if (ev.kind === "done" && looking && app.sessions[id]?.state.persisted) post(`/api/sessions/${id}/seen`).catch(() => {});
}

async function onHandoff(ev: ServerEvent) {
  const pane = app.panes.findIndex((p) => p.session === ev.session);
  if (pane < 0 || !ev.text) return;
  keep(await get<Snapshot>(`/api/sessions/${ev.text}`));
  show(ev.text, pane);
}
