export interface Trigger { kind: "/" | "#" | "@"; start: number; query: string }
export interface Item { label: string; detail?: string; insert: string; more?: boolean }

// trigger finds the completion the cursor is in: a /command at the start
// of the draft, a #prompt or an @path after whitespace.
export function trigger(text: string, cursor: number): Trigger | null {
  const before = text.slice(0, cursor);
  const command = /^\/([\w-]*)$/.exec(before);
  if (command) return { kind: "/", start: 0, query: command[1] };
  const token = /(^|\s)([#@])([^\s]*)$/.exec(before);
  if (!token) return null;
  const start = before.length - token[3].length - 1;
  return { kind: token[2] as "#" | "@", start, query: token[3] };
}

// accept puts an item in place of the trigger; items that continue (a
// folder) leave the cursor right after them.
export function accept(text: string, cursor: number, t: Trigger, item: Item): { text: string; cursor: number } {
  const insert = item.insert + (item.more ? "" : " ");
  const next = text.slice(0, t.start) + insert + text.slice(cursor).replace(/^\s/, "");
  return { text: next, cursor: t.start + insert.length };
}

// imageLabel names the n-th picture of a draft like the TUI does.
export function imageLabel(n: number): string {
  return `[image ${String(n).padStart(2, "0")}]`;
}

export function rank<T extends { label: string }>(items: T[], query: string): T[] {
  const q = query.toLowerCase();
  return items
    .filter((i) => i.label.toLowerCase().includes(q))
    .sort((a, b) => Number(!a.label.toLowerCase().startsWith(q)) - Number(!b.label.toLowerCase().startsWith(q)));
}
