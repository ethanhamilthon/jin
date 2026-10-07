import { get, query } from "./api";
import { rank, type Item, type Trigger } from "./composer";
import type { PromptInfo } from "./types";

let promptCache: { at: number; list: PromptInfo[] } | undefined;

async function promptList(): Promise<PromptInfo[]> {
  if (!promptCache || Date.now() - promptCache.at > 10_000) {
    promptCache = { at: Date.now(), list: await get<PromptInfo[]>("/api/prompts") };
  }
  return promptCache.list;
}

// items lists what may complete a trigger in a session's directory.
export async function items(t: Trigger, dir: string): Promise<Item[]> {
  if (t.kind === "#") {
    const enabled = (await promptList()).filter((p) => p.enabled);
    return rank(enabled.map((p) => ({ label: p.name, detail: p.preview, insert: "#" + p.name })), t.query).slice(0, 12);
  }
  const found = await get<{ name: string; insert: string; dir: boolean }[]>("/api/files" + query({ q: t.query, dir }));
  return found.map((f) => ({ label: f.name + (f.dir ? "/" : ""), insert: "@" + f.insert + (f.dir ? "/" : ""), more: f.dir }));
}
