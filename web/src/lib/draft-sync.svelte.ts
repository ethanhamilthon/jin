import { untrack } from "svelte";
import { app } from "./app.svelte";
import type { SessionState } from "./types";

// syncDraft keeps a composer's text per session and takes a draft the
// server puts in (a handoff brief, a rewound message).
export function syncDraft(info: () => SessionState, get: () => string, set: (text: string) => void) {
  let seen = 0;
  // A derived value tells its readers only when it changes, so the composer is
  // reset when the session changes and not on every update of its state.
  const sessionId = $derived(info().id);
  $effect(() => {
    const id = sessionId;
    untrack(() => {
      set(app.drafts[id] ?? info().draft ?? "");
      seen = info().draft_rev ?? 0;
    });
  });
  $effect(() => {
    const rev = info().draft_rev ?? 0;
    if (rev > seen) {
      seen = rev;
      set(info().draft ?? "");
    }
  });
  $effect(() => {
    app.drafts[untrack(() => info().id)] = get();
  });
}
