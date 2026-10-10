import { app, fail } from "./app.svelte";
import { post } from "./api";
import type { Kind } from "./types";

// The page has two of the TUI folding modes: tools shown (Output) and
// tools hidden (No tools). With tools hidden, reasoning is hidden too and only
// the messages and the final answer stay. The other TUI modes read as their nearest one.
export const showTools = 3;
export const hideTools = 1;

export function toolsShown(mode: number): boolean {
  return mode === 0 || mode === 3;
}

// shows says whether an entry is visible in a folding mode.
export function shows(mode: number, kind: Kind, tool?: string): boolean {
  if (tool === "task" || kind === "tool_result" || kind === "tool_call" || kind === "reasoning") return toolsShown(mode);
  return true;
}

export function setTools(shown: boolean) {
  app.config.fold = shown ? showTools : hideTools;
  post("/api/settings/fold", { fold: app.config.fold }).catch(fail);
}

export function toggleTools() {
  setTools(!toolsShown(app.config.fold));
}
