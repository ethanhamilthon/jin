export function tokens(n: number): string {
  if (n < 1000) return String(n);
  if (n < 1_000_000) return (n / 1000).toFixed(n < 10_000 ? 1 : 0).replace(/\.0$/, "") + "k";
  return (n / 1_000_000).toFixed(1).replace(/\.0$/, "") + "M";
}

export function cost(value: number): string {
  return value > 0 ? "$" + value.toFixed(4) : "—";
}

export function ago(iso: string): string {
  const seconds = (Date.now() - new Date(iso).getTime()) / 1000;
  if (seconds < 60) return "just now";
  if (seconds < 3600) return Math.floor(seconds / 60) + "m ago";
  if (seconds < 86400) return Math.floor(seconds / 3600) + "h ago";
  return Math.floor(seconds / 86400) + "d ago";
}

// shortPath writes the home folder as ~.
export function shortPath(path: string, home: string): string {
  if (!home) return path;
  if (path === home) return "~";
  return path.startsWith(home + "/") ? "~" + path.slice(home.length) : path;
}

export function shortID(id: string): string {
  return id.slice(0, 8);
}

export function percent(part: number, whole: number): number {
  return whole > 0 ? Math.round((part * 100) / whole) : 0;
}

const familyWords = new Set(["claude", "gpt", "gemini", "grok", "kimi", "glm", "qwen", "deepseek", "minimax", "codex", "flash", "lite", "pro", "mini", "nano", "plus", "high", "sol"]);
const upperWords = new Set(["gpt", "glm"]);

function capitalize(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1);
}

function prettyWord(word: string): string {
  const lower = word.toLowerCase();
  if (upperWords.has(lower)) return lower.toUpperCase();
  if (familyWords.has(lower)) return capitalize(lower);
  if (/^o\d/i.test(word)) return "O" + word.slice(1);
  return capitalize(word);
}

function joinVersions(parts: string[]): string[] {
  const out: string[] = [];
  for (const part of parts) {
    const last = out[out.length - 1];
    if (last !== undefined && /^\d+(\.\d+)*$/.test(last) && /^\d+$/.test(part)) {
      out[out.length - 1] = last + "." + part;
    } else {
      out.push(part);
    }
  }
  return out;
}

function prettyPlain(id: string): string {
  const words = joinVersions(id.split(/[-_]/).filter(Boolean)).map(prettyWord);
  if (words[0] === "GPT" && /^\d+o$/.test(words[1] ?? "")) {
    return ["GPT-" + words[1], ...words.slice(2)].join(" ");
  }
  return words.join(" ");
}

// prettyModel turns a model id such as claude-sonnet-5.5 into Claude Sonnet 5.5.
export function prettyModel(id: string): string {
  const slash = id.lastIndexOf("/");
  if (slash < 0) return prettyPlain(id);
  return id.slice(0, slash) + " · " + prettyPlain(id.slice(slash + 1));
}
