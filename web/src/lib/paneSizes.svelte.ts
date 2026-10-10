export interface LayoutSplit { col: number; row: number }

const storageKey = "jin:pane-sizes";

function clamp(value: number): number {
  return Number.isFinite(value) ? Math.min(0.85, Math.max(0.15, value)) : 0.5;
}

function load(): Record<string, LayoutSplit> {
  try {
    const data = JSON.parse(localStorage.getItem(storageKey) ?? "{}") as Record<string, LayoutSplit>;
    for (const split of Object.values(data)) {
      split.col = clamp(split.col);
      split.row = clamp(split.row);
    }
    return data;
  } catch {
    return {};
  }
}

const cache = $state<Record<string, LayoutSplit>>(load());

export function splitOf(layout: string): LayoutSplit {
  return cache[layout] ?? { col: 0.5, row: 0.5 };
}

export function setSplit(layout: string, col: number, row: number) {
  cache[layout] = { col: clamp(col), row: clamp(row) };
}

export function persistSizes() {
  try {
    localStorage.setItem(storageKey, JSON.stringify(cache));
  } catch {
    // storage may be off; sizes are then just not kept
  }
}
