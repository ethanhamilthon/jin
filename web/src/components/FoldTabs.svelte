<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { setTools, toolsShown } from "../lib/fold";

  const shown = $derived(toolsShown(app.config.fold));
</script>

<div class="tabs" role="tablist" aria-label="Tool calls in the chat (Ctrl+O)">
  <span>tools</span>
  {#each [true, false] as value (value)}
    <button role="tab" aria-selected={shown === value} class:on={shown === value} title="Ctrl+O" onclick={() => setTools(value)}>
      {value ? "show" : "hide"}
    </button>
  {/each}
</div>

<style>
  .tabs { display: flex; align-items: center; gap: 2px; font-size: 11px; color: var(--text-muted); }
  span { margin-right: 4px; }
  button {
    background: transparent; border: 1px solid transparent; border-radius: 2px; cursor: pointer;
    padding: 0 6px; color: var(--text-muted); font: inherit;
  }
  button:hover { color: var(--text-dim); }
  button.on { color: var(--accent); border-color: var(--accent-deep); }
</style>
