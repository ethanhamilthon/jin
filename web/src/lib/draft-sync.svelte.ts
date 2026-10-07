import { untrack } from "svelte";
import { app } from "./app.svelte";
import type { SessionState } from "./types";

// syncDraft keeps a composer's text per session and takes a draft the
// server puts in (a handoff brief, a rewound message).
export function syncDraft(info: () => SessionState, get: () => string, set: (text: string) => void) {
  let seen = 0;
  $effect(() => {
    const id = info().id;
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
