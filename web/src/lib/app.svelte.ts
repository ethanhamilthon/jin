import type { Config, Entry, Intro, SessionState } from "./types";

export interface SessionView { state: SessionState; entries: Entry[]; intro?: Intro; seq: number }
export interface Pane { key: number; session: string }
export interface Toast { id: number; text: string; error: boolean }
export type DialogName =
  | "model" | "providers" | "settings" | "tasks" | "context" | "rewind"
  | "undo" | "project" | "trust" | "export" | "sessions";
export interface Dialog { name: DialogName; session?: string; arg?: string }

const emptyConfig: Config = {
  providers: [], active: "", model: "", effort: "", ready: false,
  sound: { enabled: true, only_blur: false, volume: 75 }, accent: "", tools: [], scope: null, fold: 3,
};

class App {
  loaded = $state(false);
  stopped = $state(false);
  connected = $state(true);
  resyncing = false;
  version = $state("");
  dir = $state("");
  latest = $state("");
  config = $state<Config>(emptyConfig);
  sessions = $state<Record<string, SessionView>>({});
  panes = $state<Pane[]>([{ key: 1, session: "" }]);
  focused = $state(0);
  project = $state("");
  dialog = $state<Dialog | null>(null);
  toasts = $state<Toast[]>([]);
  sessionsRev = $state(0);
  projectsRev = $state(0);
  tasksRev = $state(0);
  configRev = $state(0);
  drafts: Record<string, string> = {};
  nextKey = 2;

  get pane(): Pane {
    return this.panes[Math.min(this.focused, this.panes.length - 1)];
  }

  get current(): SessionView | undefined {
    return this.sessions[this.pane.session];
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
