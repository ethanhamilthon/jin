import { app } from "./app.svelte";
import { act } from "./actions";
import { commands } from "./commands";

export interface Image { label: string; path: string }

// submit sends a draft: $ runs a shell command, a lone /command runs it,
// anything else goes to the agent with its pictures.
export async function submit(id: string, text: string, images: Image[]): Promise<boolean> {
  const trimmed = text.trim();
  if (!trimmed) return false;
  if (trimmed.startsWith("$")) return act(id, "shell", { command: trimmed.slice(1) });
  const command = /^\/([\w-]+)$/.exec(trimmed);
  if (command) {
    const found = commands.find((c) => c.name === command[1]);
    if (found) {
      found.run(id);
      return true;
    }
  }
  const used = images.filter((image) => text.includes(image.label));
  const sent = await act(id, "send", { text, images: used });
  if (sent) delete app.drafts[id];
  return sent;
}
