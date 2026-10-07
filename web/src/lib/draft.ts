import { app } from "./app.svelte";
import { act } from "./actions";

export interface Image { label: string; path: string; url?: string }
export interface AttachedFile { name: string; path: string }

// submit sends a draft: $ runs a shell command, anything else goes to the
// agent with its pictures and files.
export async function submit(id: string, text: string, images: Image[], files: AttachedFile[] = []): Promise<boolean> {
  const trimmed = text.trim();
  if (!trimmed && !images.length && !files.length) return false;
  if (trimmed.startsWith("$")) return act(id, "shell", { command: trimmed.slice(1) });
  const sent = await act(id, "send", { text, images: images.map(({ label, path }) => ({ label, path })), files });
  if (sent) delete app.drafts[id];
  return sent;
}
