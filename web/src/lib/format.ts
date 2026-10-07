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

export function baseName(path: string): string {
  const parts = path.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? path;
}

export function shortID(id: string): string {
  return id.slice(0, 8);
}

export function percent(part: number, whole: number): number {
  return whole > 0 ? Math.round((part * 100) / whole) : 0;
}
