import { app } from "./app.svelte";

export interface TreeEntry { name: string; dir: boolean; size: number }
export interface FileView { name: string; path: string; size: number; kind: "text" | "markdown" | "image" | "binary" | "large"; content?: string }

// mention is the @ text for a path, quoted when it has spaces, as the composer does.
export function mention(path: string): string {
  return /[\s"\\]/.test(path) ? `@"${path.replaceAll("\\", "\\\\").replaceAll('"', '\\"')}"` : "@" + path;
}

// insertIntoChat puts @path into the composer of the focused chat.
export function insertIntoChat(path: string) {
  const session = app.chat?.session;
  if (!session) return app.toast("Open a chat first", true);
  app.insertion = { session, text: mention(path), n: (app.insertion?.n ?? 0) + 1 };
}

export function size(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(bytes < 10240 ? 1 : 0)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

const languages: Record<string, string> = {
  go: "go", js: "javascript", mjs: "javascript", ts: "typescript", tsx: "typescript", jsx: "javascript", py: "python", rs: "rust",
  rb: "ruby", java: "java", kt: "kotlin", swift: "swift", c: "c", h: "c", cpp: "cpp", cc: "cpp", cs: "csharp", php: "php",
  sh: "bash", zsh: "bash", bash: "bash", json: "json", yaml: "yaml", yml: "yaml", toml: "ini", ini: "ini", html: "xml", xml: "xml",
  svelte: "xml", vue: "xml", css: "css", scss: "scss", sql: "sql", md: "markdown", diff: "diff", lua: "lua", dockerfile: "dockerfile",
};

// languageOf guesses the highlight.js language from a file name.
export function languageOf(name: string): string {
  const lower = name.toLowerCase();
  if (lower === "makefile") return "makefile";
  if (lower === "dockerfile") return "dockerfile";
  return languages[lower.split(".").pop() ?? ""] ?? "";
}
