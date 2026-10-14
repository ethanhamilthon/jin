import { app } from "./app.svelte";
import { mobile } from "./mobile.svelte";

// zoom shows the focused pane alone. The other panes stay open and keep running.
export const zoom = $state({ on: false });

export function zoomed(): boolean {
  return zoom.on && app.panes.length > 1 && !mobile.on;
}

export function toggleZoom() {
  if (zoom.on) {
    zoom.on = false;
    return;
  }
  if (app.panes.length > 1 && !mobile.on) zoom.on = true;
}

// blockedByZoom tells the user why a layout action does nothing.
export function blockedByZoom(): boolean {
  if (!zoomed()) return false;
  app.toast("Exit zoom first");
  return true;
}
