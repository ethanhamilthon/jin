export type Kind =
  | "user" | "assistant" | "reasoning" | "tool_call" | "tool_result" | "info" | "error"
  | "tell" | "ask" | "todo" | "compacted";

export interface Line { op: " " | "+" | "-"; text: string }
export interface Entry { kind: Kind; tool?: string; text: string; lines?: Line[] }
export interface Usage { Input: number; Output: number; Context: number; Cost: number }
export interface TodoItem { text: string; status: "pending" | "in_progress" | "done" }
export interface Question { question: string; options?: string[]; multiple?: boolean }

export interface SessionState {
  id: string; path: string; title: string; model: string; effort: string; provider: string;
  persisted: boolean; ready: boolean; working: boolean; busy: boolean; unread: boolean;
  read_only?: number; provider_missing?: boolean; usage: Usage; cache?: number; window: number;
  todos: TodoItem[]; ask?: Question[]; suggestion?: string; loading?: string[]; reloading?: boolean;
  shell?: boolean; queued?: number; tasks?: number; draft?: string; draft_rev?: number;
}

export interface Intro {
  version: string; tools: string[]; context: string[] | null; hooks: string[] | null;
  prompts: string[] | null; system_prompt?: string; update?: string;
}

export interface Snapshot { state: SessionState; entries: Entry[]; intro?: Intro; seq: number }

export interface Provider { id: string; name: string; kind: string; base_url: string }
export interface Tool { name: string; description: string; enabled: boolean }
export interface Sound { enabled: boolean; only_blur: boolean; volume: number }

export interface Config {
  providers: Provider[]; active: string; model: string; effort: string; ready: boolean;
  sound: Sound; accent: string; tools: Tool[]; scope: string[] | null;
}

export interface AppState { version: string; dir: string; latest: string; config: Config; live: SessionState[] | null }

export interface SessionRow {
  id: string; title: string; path: string; model: string; updated_at: string;
  usage: Usage; unread: boolean; snippet?: string;
}

export interface Project {
  id: string; path: string; name: string; last_session?: string; sessions: number; unread: boolean;
}

export interface Model { id: string; context?: number; input?: number; output?: number; reasoning?: boolean; vision?: boolean }
export interface PromptInfo { name: string; system: boolean; preview: string; enabled: boolean }
export interface Hook { name: string; project: boolean; dir: string; enabled: boolean; preview: string }
export interface Task { id: string; owner: string; command: string; dir: string; status: string; exit: number; started: string }
export interface Part { name: string; tokens: number }

export interface ContextReport {
  used: number; window: number; prompt: Part[] | null; tool_schemas: number;
  messages: number; conversation: number; results: Part[] | null; cache?: number;
}

export interface Point { index: number; text: string }

export interface ServerEvent {
  type: string; session?: string; index?: number; entry?: Entry; entries?: Entry[];
  text?: string; kind?: string; state?: SessionState; seq: number;
}
