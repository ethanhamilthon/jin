import { app, fail } from "./app.svelte";
import { post } from "./api";
import type { Kind } from "./types";

// The folding modes of the TUI, in its Ctrl+O order; the values are the
// ones the TUI saves.
export const folds = [
  { value: 3, label: "Output", title: "Tool calls and their output" },
  { value: 0, label: "All", title: "Tool calls without their output" },
  { value: 1, label: "No tools", title: "Messages and reasoning" },
  { value: 2, label: "Messages", title: "Messages only" },
];

// shows says whether an entry kind is visible in a folding mode.
export function shows(mode: number, kind: Kind): boolean {
  if (kind === "tool_result") return mode === 3;
  if (kind === "tool_call") return mode === 0 || mode === 3;
  if (kind === "reasoning") return mode !== 2;
  return true;
}

export function setFold(mode: number) {
  app.config.fold = mode;
  post("/api/settings/fold", { fold: mode }).catch(fail);
}

export function nextFold() {
  const i = folds.findIndex((f) => f.value === app.config.fold);
  setFold(folds[(i + 1) % folds.length].value);
}
