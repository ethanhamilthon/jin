import { app, type Pane } from "./app.svelte";
import { post } from "./api";
import { keep } from "./actions";
import type { Snapshot } from "./types";

const storageKey = "jin.layout";

interface Saved { panes: { kind: Pane["kind"]; session: string; section?: string; file?: string }[]; focused: number }

// saveLayout remembers the open panes of this browser.
export function saveLayout() {
  const saved: Saved = {
    panes: app.panes.map((p) => ({ kind: p.kind, session: p.session, section: p.section, file: p.file })),
    focused: app.focused,
  };
  try {
    localStorage.setItem(storageKey, JSON.stringify(saved));
  } catch {
    // storage may be off; the layout is then just not kept
  }
}

// restoreLayout opens the panes of the last visit. A chat whose session is gone
// or taken by another process is dropped; at least one chat is always left.
export async function restoreLayout() {
  let saved: Saved | null = null;
  try {
    saved = JSON.parse(localStorage.getItem(storageKey) ?? "null");
  } catch {
    saved = null;
  }
  if (!saved?.panes?.length) return;
  const panes: Pane[] = [];
  for (const entry of saved.panes.slice(0, 4)) {
    if (entry.kind === "chat") {
      if (!entry.session) continue;
      try {
        keep(await post<Snapshot>(`/api/sessions/${entry.session}/open`));
      } catch {
        continue;
      }
    }
    panes.push({ key: app.nextKey++, kind: entry.kind, session: entry.session, section: entry.section, file: entry.file });
  }
  if (!panes.some((p) => p.kind === "chat")) return;
  app.panes = panes;
  app.focused = Math.min(Math.max(saved.focused, 0), panes.length - 1);
  app.project = app.current?.state.path ?? app.project;
}

// saveSidebar remembers whether the sidebar is open.
export function saveSidebar(open: boolean) {
  try {
    localStorage.setItem("jin.sidebar", open ? "1" : "0");
  } catch {
    // not kept
  }
}
