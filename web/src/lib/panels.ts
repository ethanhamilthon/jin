import { app, type PaneKind } from "./app.svelte";

export type PanelKind = Exclude<PaneKind, "chat">;

// openPanel shows a panel. There is one pane of each kind: it is focused if it
// is open. Otherwise it takes a free place, or the place of the focused pane
// when four panes are open; a chat that loses its pane keeps running.
export function openPanel(kind: PanelKind, section?: string) {
  const open = app.panes.findIndex((p) => p.kind === kind);
  if (open >= 0) {
    if (section !== undefined) app.panes[open].section = section;
    app.focused = open;
    return;
  }
  const pane = { key: app.nextKey++, kind, session: "", section };
  if (app.panes.length < 4) {
    app.panes.push(pane);
    app.focused = app.panes.length - 1;
    return;
  }
  app.panes[app.focused] = pane;
  app.focused = app.focused;
}

// closeSection goes back from a Settings section to the list of sections.
export function closeSection() {
  const pane = app.panes.find((p) => p.kind === "settings");
  if (pane) pane.section = undefined;
}
