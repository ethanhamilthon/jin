import { app } from "./app.svelte";
import { act, closePane, newSession, split } from "./actions";
import { get } from "./api";

export interface Command { name: string; description: string; run: (id: string) => void }

async function exportMarkdown(id: string) {
  const { markdown } = await get<{ markdown: string }>(`/api/sessions/${id}/export`);
  const link = document.createElement("a");
  link.href = URL.createObjectURL(new Blob([markdown], { type: "text/markdown" }));
  link.download = `jin-${id.slice(0, 8)}.md`;
  link.click();
  URL.revokeObjectURL(link.href);
}

// commands are the actions of the / list and the command palette.
export const commands: Command[] = [
  { name: "new", description: "New session in this project", run: () => newSession(app.current?.state.path) },
  { name: "split", description: "Open another pane with a new session", run: split },
  { name: "close", description: "Close the focused pane", run: () => closePane() },
  { name: "model", description: "Select the model and effort", run: (id) => app.open("model", id) },
  { name: "provider", description: "Providers: add, switch, delete", run: () => app.open("providers") },
  { name: "project", description: "Add a project directory", run: () => app.open("project") },
  { name: "settings", description: "Sound, accent, tools, prompts, hooks, data folder", run: () => app.open("settings") },
  { name: "tasks", description: "Background tasks: output and stop", run: () => app.open("tasks") },
  { name: "context", description: "Show what fills the context", run: (id) => app.open("context", id) },
  { name: "compact", description: "Summarize the conversation to free context", run: (id) => act(id, "compact") },
  { name: "handoff", description: "Continue the work in a fresh session", run: (id) => act(id, "handoff") },
  { name: "rewind", description: "Restart the conversation from a message; files stay", run: (id) => app.open("rewind", id) },
  { name: "undo", description: "Restore edit and write changes of the last turn", run: (id) => app.open("undo", id) },
  { name: "reload", description: "Reload this session's prompts, hooks and instructions", run: (id) => act(id, "reload") },
  { name: "stop", description: "Interrupt the shell command or request", run: (id) => act(id, "stop") },
  { name: "export", description: "Download the session as Markdown", run: (id) => exportMarkdown(id).catch((e) => app.toast(String(e), true)) },
];
