import type { Provider } from "./types";

export const isEnabled = (provider: Provider) => provider.enabled !== false;

// noEnabledProvider is true when a prompt has no enabled provider to go to: the
// list has providers and none is enabled, or the session's own provider is off.
export function noEnabledProvider(providers: Provider[], session: string): boolean {
  if (!providers.length) return false;
  if (!providers.some(isEnabled)) return true;
  return providers.some((p) => p.id === session && !isEnabled(p));
}
