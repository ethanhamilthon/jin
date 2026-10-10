import type { Config, Entry, Intro, SessionState } from "./types";

export interface SessionView { state: SessionState; entries: Entry[]; intro?: Intro; seq: number }
export type PaneKind = "chat" | "settings" | "files" | "project";
// A pane is a chat (with a session) or a panel. section is the open Settings
// section, file the file open in Files, project the project a Project pane is
// pinned to (empty: the project of the focused chat).
export interface Pane { key: number; kind: PaneKind; session: string; section?: string; file?: string; project?: string }
export interface Toast { id: number; text: string; error: boolean }
export type DialogName =
  | "model" | "providers" | "tasks" | "context" | "rewind"
  | "undo" | "addproject" | "trust" | "export" | "switch" | "pair";
export interface Dialog { name: DialogName; session?: string; arg?: string }

const emptyConfig: Config = {
  providers: [], active: "", model: "", effort: "", ready: false,
  sound: { enabled: true, only_blur: false, volume: 75 }, accent: "", tools: [], scope: null, fold: 3,
};

class App {
  loaded = $state(false);
  restored = $state(false);
  stopped = $state(false);
  connected = $state(true);
  resyncing = false;
  version = $state("");
  dir = $state("");
  home = $state("");
  latest = $state("");
  config = $state<Config>(emptyConfig);
  sessions = $state<Record<string, SessionView>>({});
  panes = $state<Pane[]>([{ key: 1, kind: "chat", session: "" }]);
  #focused = $state(0);
  lastChat = $state(1);
  // insertion is text a panel puts into the composer of a chat (n tells a new one).
  insertion = $state<{ session: string; text: string; n: number } | null>(null);
  project = $state("");
  dialog = $state<Dialog | null>(null);
  toasts = $state<Toast[]>([]);
  sessionsRev = $state(0);
  projectsRev = $state(0);
  tasksRev = $state(0);
  configRev = $state(0);
  drafts: Record<string, string> = {};
  nextKey = 2;

  get focused(): number {
    return Math.min(this.#focused, this.panes.length - 1);
  }

  set focused(index: number) {
    this.#focused = index;
    const pane = this.panes[index];
    if (pane?.kind === "chat") this.lastChat = pane.key;
  }

  get pane(): Pane {
    return this.panes[this.focused];
  }

  // chat is the pane that chat actions go to: the focused one when it is a
  // chat, else the chat that was focused last.
  get chat(): Pane | undefined {
    if (this.pane.kind === "chat") return this.pane;
    return this.panes.find((p) => p.key === this.lastChat && p.kind === "chat") ?? this.panes.find((p) => p.kind === "chat");
  }

  get chatIndex(): number {
    const chat = this.chat;
    return chat ? this.panes.indexOf(chat) : 0;
  }

  get current(): SessionView | undefined {
    return this.sessions[this.chat?.session ?? ""];
  }

  // projectPath is the project of the chat in focus, for the Files and Project panes.
  get projectPath(): string {
    return this.current?.state.path || this.project || this.dir;
  }

  visible(id: string): boolean {
    return this.panes.some((p) => p.session === id);
  }

  live(): SessionState[] {
    return Object.values(this.sessions).map((s) => s.state);
  }

  toast(text: string, error = false) {
    const id = Date.now() + Math.random();
    this.toasts.push({ id, text, error });
    setTimeout(() => (this.toasts = this.toasts.filter((t) => t.id !== id)), error ? 7000 : 3500);
  }

  open(name: DialogName, session?: string, arg?: string) {
    this.dialog = { name, session: session || this.pane.session, arg };
  }
}

export const app = new App();

export function fail(err: unknown) {
  app.toast(err instanceof Error ? err.message : String(err), true);
}
