async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const init: RequestInit = { method, headers: {} };
  if (body instanceof Blob) {
    init.body = body;
  } else if (body !== undefined) {
    init.body = JSON.stringify(body);
    init.headers = { "Content-Type": "application/json" };
  }
  const response = await fetch(path, init);
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error ?? response.statusText);
  return data as T;
}

export const get = <T>(path: string) => request<T>("GET", path);
export const post = <T = unknown>(path: string, body?: unknown) => request<T>("POST", path, body ?? {});

export function query(params: Record<string, string | undefined>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) if (value) search.set(key, value);
  const text = search.toString();
  return text ? "?" + text : "";
}
