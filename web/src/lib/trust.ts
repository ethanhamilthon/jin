import { app } from "./app.svelte";
import { get, query } from "./api";
import type { Hook } from "./types";

const asked = new Set<string>();

// askTrust asks once per page and project whether its own hooks may run,
// as long as nobody answered yet.
export async function askTrust(dir: string) {
  if (!dir || asked.has(dir)) return;
  asked.add(dir);
  const { hooks, trust } = await get<{ hooks: Hook[]; trust: number }>("/api/hooks" + query({ dir }));
  if (trust === 0 && hooks.some((h) => h.project) && !app.dialog) app.open("trust", undefined, dir);
}
