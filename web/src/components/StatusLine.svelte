<script lang="ts">
  import { app } from "../lib/app.svelte";
  import type { SessionState } from "../lib/types";
  import { percent } from "../lib/format";
  import Icon from "./Icon.svelte";
  import FoldTabs from "./FoldTabs.svelte";

  let { state }: { state: SessionState } = $props();
  const share = $derived(percent(state.usage.Context, state.window));
</script>

<div class="status mono">
  <button class="context" class:warn={share >= 70} title="Context: window, spent, input and output" aria-label="Context" onclick={() => app.open("context", state.id)}><Icon name="context" size={14} /></button>
  {#if state.queued}<span class="queued">{state.queued} queued</span>{/if}
  <FoldTabs />
</div>

<style>
  .status { flex: 1; display: flex; align-items: center; gap: 12px; flex-wrap: wrap; font-size: 11px; color: var(--text-muted); padding: 0 2px; }
  .context {
    background: transparent; border: 1px solid transparent; border-radius: 2px; cursor: pointer;
    padding: 3px 6px; display: inline-flex; color: var(--text-muted);
  }
  .context:hover { color: var(--text-dim); border-color: var(--line); }
  .context.warn { color: var(--warn); }
  .queued { color: var(--accent); }
</style>
