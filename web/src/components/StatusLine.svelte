<script lang="ts">
  import type { SessionState } from "../lib/types";
  import { cost, percent, tokens } from "../lib/format";
  import FoldTabs from "./FoldTabs.svelte";

  let { state }: { state: SessionState } = $props();
  const share = $derived(percent(state.usage.Context, state.window));
</script>

<div class="status mono">
  <span title="Input tokens">↑ {tokens(state.usage.Input)}</span>
  <span title="Output tokens">↓ {tokens(state.usage.Output)}</span>
  <span class:warn={share >= 70} title="Context">
    ◫ {tokens(state.usage.Context)}{#if state.window}/{tokens(state.window)} · {share}%{/if}
  </span>
  {#if state.cache !== undefined && state.cache !== null}<span title="Cached input of the last request">cache {state.cache}%</span>{/if}
  <span title="Cost">{cost(state.usage.Cost)}</span>
  {#if state.queued}<span class="queued">{state.queued} queued</span>{/if}
  <span class="spacer"></span>
  <FoldTabs />
</div>

<style>
  .status { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; font-size: 11px; color: var(--text-muted); padding: 0 2px; }
  .warn { color: var(--warn); }
  .spacer { flex: 1; }
  .queued { color: var(--accent); }
</style>
